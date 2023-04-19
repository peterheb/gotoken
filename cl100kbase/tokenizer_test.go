// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package cl100kbase_test

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/peterheb/gotoken"
	"github.com/peterheb/gotoken/cl100kbase"
	"github.com/peterheb/gotoken/internal"
)

const (
	testInput     = "../testdata/samples.txt"
	testExpected  = "../testdata/cl100k_base.txt"
	oneGBinput    = "../testdata/pae-enwiki-2023-04-1gb.txt"
	oneGBexpected = "../testdata/cl100k_1gb.txt"
)

func TestTokenizer(t *testing.T) {
	var tpr *internal.TestPairReader
	var err error
	if len(os.Getenv("GOTOKEN_TEST_1GB")) > 0 {
		tpr, err = internal.NewTestPairReader(oneGBinput, oneGBexpected)
	} else {
		tpr, err = internal.NewTestPairReader(testInput, testExpected)
	}
	if err != nil {
		t.Fatalf("loading test data: %v", err)
	}
	defer tpr.Close()

	tok, err := gotoken.GetTokenizer("cl100k_base")
	if err != nil {
		t.Fatalf("instantiating tokenizer: %v", err)
	}

	end := false
	for {
		tc, err := tpr.Next()
		if tc == nil && err == nil {
			// Append one final test case to cover trailing newlines, which due
			// to our textual samples.txt file format, is not otherwise tested.
			tc = &internal.TestPair{
				Input:    "Hello world.\r\n\r\n",
				Expected: []int{9906, 1917, 18304},
			}
			end = true
		}
		if err != nil {
			t.Fatalf("loading test data: %v", err)
		}
		t.Run(tpr.CaseName(), func(t *testing.T) {
			actual, err := tok.Encode(tc.Input)
			if err != nil {
				t.Errorf("Encode(%q): %v", tc.Input, err)
			}
			decoded, err := tok.Decode(actual)
			if err != nil {
				t.Errorf("Decode(%#v): %v", actual, err)
			}
			if decoded != tc.Input {
				t.Errorf("Decode(%#v) got %q; expected %q", actual, decoded, tc.Input)
			}
			if !reflect.DeepEqual(actual, tc.Expected) {
				t.Errorf("Encode(%q) got %#v; expected %#v", tc.Input, actual, tc.Expected)
			}
		})
		if end {
			break
		}
	}
}

func FuzzCL100K(f *testing.F) {
	tok, err := gotoken.GetTokenizer("cl100k_base", gotoken.WithSpecialTokensAsText())
	if err != nil {
		f.Fatalf("instantiating tokenizer: %v", err)
	}

	// Seed the corpus with some interesting strings including letters, numbers,
	// whitespace patterns, unicode, emoji, broken UTF-8, and a special token.
	seedCorpus := []string{
		"", " ", "\r\n", "a", "a a", "abc:\r\n    1 23 (456) 7,890.12\r\n\r\n",
		"Hello, <b>world</b>!", "안녕, 세상!", "It's done! 🎉", "a://b x=`    `;",
		"x's x'lL.\r\n x'vE", "\xc0\xff\xed\xa0\x80", "a" + cl100kbase.EndOfText + "b",
	}
	for _, s := range seedCorpus {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, orig string) {
		// For each iteration, verify that Encode() and Decode() complete
		// without errors and that x == Decode(Encode(x))
		tokens, err := tok.Encode(orig)
		if err != nil {
			t.Fatalf("Encode(%q): %v", orig, err)
		}
		rev, err := tok.Decode(tokens)
		if err != nil {
			t.Fatalf("Decode(%#v): %v", tokens, err)
		}

		if !bytes.Equal([]byte(orig), []byte(rev)) {
			t.Errorf("Before: %#v, after: %#v", []byte(orig), []byte(rev))
		}
	})
}
