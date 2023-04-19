// Copyright 2023 Peter Hebert. Licensed under the MIT license.

// Package gen generates the data.go files for the gotoken's encoding
// sub-packages. To run, "go generate". This will overwrite files in
// ../{encoding}/data.go!
//
// This program generates data structures based on the ".tiktoken" files
// provided by OpenAI. These files are licensed under the MIT license and
// include the following notice: Copyright (c) 2022 OpenAI, Shantanu Jain.
//
//   - See also: https://github.com/openai/tiktoken/blob/main/LICENSE
//
//go:generate go run gen.go trie.go
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/peterheb/gotoken/internal"
)

func main() {
	generate("r50k_base", "https://openaipublic.blob.core.windows.net/encodings/r50k_base.tiktoken")
	generate("p50k_base", "https://openaipublic.blob.core.windows.net/encodings/p50k_base.tiktoken")
	generate("cl100k_base", "https://openaipublic.blob.core.windows.net/encodings/cl100k_base.tiktoken")
}

func generate(encoding, src string) {
	encodingPkg := strings.ReplaceAll(encoding, "_", "")
	outFilename := fmt.Sprintf("../%s/data.go", encodingPkg)
	dir, err := os.Stat("../" + encodingPkg)
	onErrFatalf(err, "stat '../%s': %v (are you running this from the gen/ folder?)", encodingPkg, err)
	assert(dir.IsDir(), "'../%s' is not a directory", encodingPkg)

	fmt.Printf("retrieving %s... ", src)
	contents, err := readFileFromURL(src)
	onErrFatalf(err, "loading source data")
	fmt.Println("OK")

	// decode the input file into a map first to get our bearings
	fmt.Print("decoding... ")
	lines := strings.Split(string(contents), "\n")
	tokens := make(map[int]string)
	byteTokens := make([]int, 256)
	lo, hi := 0xfffffff, -1
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}

		fields := strings.Fields(line)
		assert(len(fields) == 2, "expected 2 fields, got %d on line %d: %q", len(fields), i, line)
		token, err := base64.StdEncoding.DecodeString(fields[0])
		onErrFatalf(err, "decoding base64 '%s'", fields[0])
		rank, err := strconv.Atoi(fields[1])
		onErrFatalf(err, "strconv.Atoi('%s')", fields[1])
		tokens[rank] = string(token)
		if len(token) == 1 {
			// sanity check: if this is not true, we need to rethink our data structures!
			assert(rank < 256, "expected single byte token %q to have rank<256, got %d", token, rank)
			byteTokens[token[0]] = rank
		}
		lo = min(lo, rank)
		hi = max(hi, rank)
	}

	// convert the map to a slice
	assert(lo == 0, "lo == %d, expected 0", lo)
	allTokens := make([]string, hi+1)
	for k, v := range tokens {
		allTokens[k] = v
	}

	// allTokens = allTokens[:512]    // uncomment to make the data in baby_test.go
	bytePairLookup := createPairList(allTokens)
	fmt.Println("OK")

	fmt.Print("building trie... ")
	trie := BuildTrie(allTokens)
	serialized := trie.serialize()
	assert(serialized[0]&0xff == 0, "trie root node is not 256-ary: got %d", serialized[0]>>8)
	// verify serialized trie integrity by looking up every token
	for i, token := range allTokens {
		if len(token) == 0 {
			continue
		}
		lkup := internal.TrieLookup(serialized, []byte(token))
		assert(i == lkup, "trie build failure: lookup(%q): wanted=%d got=%d\n", token, i, lkup)
	}
	fmt.Printf("OK (%d nodes)\n", len(serialized))

	fmt.Printf("creating ../%s/data.go... ", encodingPkg)
	f := &bytes.Buffer{}
	fmt.Fprint(f, "// Code generated programmatically by go generate; DO NOT EDIT\n\n")
	fmt.Fprintf(f, "// Package %s registers the %#v tokenizer with gotoken.\n", encodingPkg, encoding)
	fmt.Fprintln(f, "// To use this tokenizer:")
	fmt.Fprintln(f, "//")
	fmt.Fprintln(f, "//   import (")
	fmt.Fprintln(f, `//       "github.com/peterheb/gotoken"`)
	fmt.Fprintf(f, "//       _ \"github.com/peterheb/gotoken/%s\"\n", encodingPkg)
	fmt.Fprintln(f, "//   )")
	fmt.Fprintln(f, "//   ...")
	fmt.Fprintf(f, "//   tok, err := gotoken.GetTokenizer(%#v)\n", encoding)
	fmt.Fprintln(f, "//")
	fmt.Fprintln(f, "// This file was generated from the following data:")
	fmt.Fprintln(f, "//")
	fmt.Fprintf(f, "//   - Source URL: %s\n", src)
	fmt.Fprintf(f, "//   - Source SHA-256: %s\n", calcSHA256(contents))
	fmt.Fprintf(f, "//   - Generated: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(f, "package %s\n", encodingPkg)
	fmt.Fprintln(f)
	fmt.Fprintln(f, "// byteToToken translates raw bytes to their token values")
	fmt.Fprintln(f, "var byteToToken = []byte{")
	emitSlice(f, byteTokens, formatted("%d,"))
	fmt.Fprintln(f, "\n}")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "// tokenList is the full list of tokens as strings, for decoding")
	fmt.Fprintln(f, "var tokenList = []string{")
	emitSlice(f, allTokens, formatted("%q,"))
	fmt.Fprintln(f, "\n}")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "// tokenTrie is a serialized map[string]int of token string -> rank")
	fmt.Fprintln(f, "var tokenTrie = []uint32{")
	emitSlice(f, serialized, hexOrDigit)
	fmt.Fprintln(f, "\n}")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "// bytePairLookup maps pairs of bytes to tokens, for kicking off BPE")
	fmt.Fprintln(f, "var bytePairLookup = []int64{")
	emitSlice(f, bytePairLookup, hexOrDigit)
	fmt.Fprintln(f, "\n}")

	// format the code we just wrote with "go fmt"
	code := f.Bytes()
	formatted, err := format.Source(code)
	if err != nil {
		err2 := os.WriteFile("./broken.txt", code, 0644)
		onErrFatalf(err2, "writing broken.txt")
	}
	onErrFatalf(err, "formatting output (see ./broken.txt)")

	// and save
	err = os.WriteFile(outFilename, formatted, 0644)
	onErrFatalf(err, "writing output file")
	fmt.Printf("wrote %d bytes\n", len(formatted))
}

// readFileFromURL loads the contents of a URL into a byte slice.
func readFileFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return data, nil
}

// This style of error handling is a little weird for Go programs, but since
// this is a CLI program that does one task, it's appropriate to just fatal on
// an error if we can't recover. As a plus, it's fewer lines than if err { ... }
// over and over and over...

// onErrFatalf prints a message and ends the program if err!=nil.
func onErrFatalf(err error, format string, args ...any) {
	if err != nil {
		fmt.Printf(format, args...)
		fmt.Printf(": %v\n", err)
		os.Exit(1)
	}
}

// assert prints a message and ends the program if cond==false.
func assert(cond bool, format string, args ...any) {
	if !cond {
		fmt.Print("assertion failed: ")
		fmt.Printf(format, args...)
		os.Exit(1)
	}
}

// emitSlice emits a slice of any type, one element per line, using format
// string "format" and a maximum line length of 72-ish characters. We don't
// count the leading tab in that 72, hence the "ish".
func emitSlice(f io.Writer, data any, formatter func(v reflect.Value) string) {
	val := reflect.ValueOf(data)
	assert(val.Kind() == reflect.Slice, "emitSlice: not a slice")
	chars := 0
	for i := 0; i < val.Len(); i++ {
		valueStr := formatter(val.Index(i))
		if chars+len(valueStr) > 72 {
			fmt.Fprint(f, "\n")
			chars = 0
		}
		if chars > 0 {
			fmt.Fprint(f, " ")
			chars++
		}
		fmt.Fprint(f, valueStr)
		chars += len(valueStr)
	}
}

// formatted returns a func that formats v with fstr.
func formatted(fstr string) func(v reflect.Value) string {
	return func(v reflect.Value) string {
		return fmt.Sprintf(fstr, v)
	}
}

// hexOrDigit formats v as a hex string if it's >=10, or just emits a single
// digit otherwise. It also outputs a comma.
func hexOrDigit(v reflect.Value) string {
	val := 0
	switch v.Kind() {
	case reflect.Uint32:
		val = int(v.Uint())
	case reflect.Int:
		val = int(v.Int())
	}
	if val < 10 {
		return fmt.Sprintf("%d,", val)
	}
	return fmt.Sprintf("0x%x,", val)
}

// calcSHA256 returns the SHA256 of a []byte as a hex string.
func calcSHA256(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

// createPairList returns a list of left<<28|right<<20|n for all two-character
// tokens. left and right are the string byte values.
func createPairList(tokenList []string) []int {
	pairList := make([]int, 0)
	byteToToken := make([]int, 256)
	for i := 0; i < 256; i++ {
		byteToToken[tokenList[i][0]] = i
	}

	for i := 256; i < len(tokenList); i++ {
		if len(tokenList[i]) == 2 {
			pairList = append(pairList, (int(tokenList[i][0])<<28)|(int(tokenList[i][1])<<20)|i)
		}
	}

	sort.Ints(pairList)
	return pairList
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
