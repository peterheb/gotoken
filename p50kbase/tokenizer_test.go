// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package p50kbase_test

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/peterheb/gotoken"
	"github.com/peterheb/gotoken/internal"
	"github.com/peterheb/gotoken/p50kbase"
)

// p50k_base and p50k_edit only differ in their special token list; the text
// tokenization is the same for both.

const (
	testInput     = "../testdata/samples.txt"
	testExpected  = "../testdata/p50k_base.txt"
	oneGBinput    = "../testdata/pae-enwiki-2023-04-1gb.txt"
	oneGBexpected = "../testdata/p50k_1gb.txt"
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

	tokName := []string{"p50k_base", "p50k_edit"}
	toks := make([]gotoken.Tokenizer, 0, 2)
	for _, encoding := range tokName {
		tok, err := gotoken.GetTokenizer(encoding)
		if err != nil {
			t.Fatalf("instantiating tokenizer %q: %v", encoding, err)
		}
		toks = append(toks, tok)
	}

	for {
		tc, err := tpr.Next()
		if tc == nil && err == nil {
			break
		}
		if err != nil {
			t.Fatalf("loading test data: %v", err)
		}
		t.Run(tpr.CaseName(), func(t *testing.T) {
			for j, tok := range toks {
				actual, err := tok.Encode(tc.Input)
				if err != nil {
					t.Errorf("%s.Encode(%q): %v", tokName[j], tc.Input, err)
				}
				decoded, err := tok.Decode(actual)
				if err != nil {
					t.Errorf("%s.Decode(%#v): %v", tokName[j], actual, err)
				}
				if decoded != tc.Input {
					t.Errorf("%s.Decode(%#v) got %q; expected %q", tokName[j], actual, decoded, tc.Input)
				}
				if !reflect.DeepEqual(actual, tc.Expected) {
					t.Errorf("%s.Encode(%q) got %#v; expected %#v", tokName[j], tc.Input, actual, tc.Expected)
				}
			}
		})
	}
}

func FuzzP50K(f *testing.F) {
	tok, err := gotoken.GetTokenizer("p50k_base", gotoken.WithSpecialTokensAsText())
	if err != nil {
		f.Fatalf("instantiating tokenizer: %v", err)
	}

	// Seed the corpus with some interesting strings including letters, numbers,
	// whitespace patterns, unicode, emoji, broken UTF-8, and a special token.
	seedCorpus := []string{
		"", " ", "\r\n", "a", "a a", "abc:\r\n    1 23 (456) 7,890.12\r\n\r\n",
		"Hello, <b>world</b>!", "안녕, 세상!", "It's done! 🎉", "a://b x=`    `;",
		"x's x'll.\r\n I've", "\xc0\xff\xed\xa0\x80", "a" + p50kbase.EndOfText + "b",
	}
	for _, s := range seedCorpus {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// For each iteration, verify that Encode() and Decode() complete
		// without errors and that x == Decode(Encode(x))
		tokens, err := tok.Encode(input)
		if err != nil {
			t.Fatalf("Encode(%q): %v", input, err)
		}
		decoded, err := tok.Decode(tokens)
		if err != nil {
			t.Fatalf("Decode(%#v): %v", tokens, err)
		}

		// Fuzz test fails if the input didn't round-trip!
		if !bytes.Equal([]byte(input), []byte(decoded)) {
			t.Errorf("Before: %#v\nTokenized: %#v\nDecoded: %#v", []byte(input), tokens, []byte(decoded))
		}
	})
}
