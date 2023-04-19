// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package r50kbase_test

import (
	"bufio"
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/peterheb/gotoken"
	"github.com/peterheb/gotoken/internal"
	"github.com/peterheb/gotoken/r50kbase"
)

const (
	testInput     = "../testdata/samples.txt"
	testExpected  = "../testdata/r50k_base.txt"
	oneGBinput    = "../testdata/pae-enwiki-2023-04-1gb.txt"
	oneGBexpected = "../testdata/r50k_1gb.txt"
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

	tok, err := gotoken.GetTokenizer("r50k_base")
	if err != nil {
		t.Fatalf("instantiating tokenizer: %v", err)
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
	}
}

func FuzzR50K(f *testing.F) {
	tok, err := gotoken.GetTokenizer("r50k_base", gotoken.WithSpecialTokensAsText())
	if err != nil {
		f.Fatalf("instantiating tokenizer: %v", err)
	}

	// Seed the corpus with some interesting strings including letters, numbers,
	// whitespace patterns, unicode, emoji, broken UTF-8, and a special token.
	seedCorpus := []string{
		"", " ", "\r\n", "a", "a a", "abc:\r\n    1 23 (456) 7,890.12\r\n\r\n",
		"Hello, <b>world</b>!", "안녕, 세상!", "It's done! 🎉", "a://b x=`    `;",
		"x's x'll.\r\n I've", "\xc0\xff\xed\xa0\x80", "a" + r50kbase.EndOfText + "b",
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

		if !bytes.Equal([]byte(input), []byte(decoded)) {
			t.Errorf("Before: %#v\nTokenized: %#v\nDecoded: %#v", []byte(input), tokens, []byte(decoded))
		}
	})
}

func BenchmarkEncode(b *testing.B) {
	f, err := os.Open(oneGBinput)
	if err != nil {
		b.Skip("skipping encode benchmark because "+oneGBinput+" could not be opened", err)
		return
	}
	defer f.Close()

	tok, err := gotoken.GetTokenizer("r50k_base")
	if err != nil {
		b.Fatalf("instantiating tokenizer: %v", err)
	}

	// Run the benchmark
	b.ResetTimer()

	// Process the input file as many times as necessary
	i := 0
	for i < b.N {
		rdr := bufio.NewReaderSize(f, 16*1024*1024)
		scan := bufio.NewScanner(rdr)
		scan.Split(bufio.ScanLines)
		for scan.Scan() {
			line := scan.Bytes()
			i++
			_, err := tok.Encode(string(line))
			if err != nil {
				b.Fatalf("Encode[line=%d](%q): %v", i, string(line), err)
			}
			if i >= b.N {
				break
			}
		}
		f.Close()
		if i < b.N {
			f, err = os.Open(oneGBinput)
			if err != nil {
				b.Fatalf("re-opening file: %v", err)
			}
			// Yes, I know some of these will be excessive defer close calls.
			// We're just reading. An error is harmless.
			defer f.Close()
		}
	}
}
