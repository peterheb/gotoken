// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package internal

import (
	"fmt"
	"reflect"
	"testing"
)

type bpeTest struct {
	text   string
	tokens []int
}

func must(t *testing.T, cond bool, msg string, params ...any) {
	t.Helper()
	if !cond {
		t.Fatalf(msg, params...)
	}
}

// instantiate the "baby" tokenizer (r50k_base, truncated to 512 tokens) with
// special tokens enabled
func getBabyBPETokenizer(allowSpecialAsText bool, allowedSpecial []string) (*BPETokenizer, error) {
	return NewBPETokenizer(getBabyTokenizerParams(), allowSpecialAsText, allowedSpecial)
}

func TestNewBPETokenizer(t *testing.T) {
	// Validate the fields of the BPETokenizer struct
	bpe, err := getBabyBPETokenizer(true, []string{})
	must(t, err == nil, "NewBPETokenizer: %v", err)
	must(t, bpe.disallowSpecialTokens == false, "bpe.disallowSpecialTokens != false")
	must(t, len(bpe.allowedSpecialTokens) == 0, "len(bpe.allowedSpecialTokens) > 0")
	must(t, bpe.decodeSpecialTokens[babyEndOfTextToken] == babyEndOfTextString, "bpe.decodeSpecialTokens[] not initialized correctly")
	must(t, bpe.params.Name == "baby", "bpe.params not correct")
	must(t, bpe.specialTokenRegex != nil, "bpe.specialTokenRegex not set")

	// Make sure NewBPETokenizer rejects invalid special tokens on the allow
	// list
	_, err = NewBPETokenizer(getBabyTokenizerParams(), false, []string{"<|not_special|>"})
	must(t, err != nil, "NewBPETokenizer: did not reject bad special token in allow list")

	// Test the bpe.specialTokenRegex created by NewBPETokenizer
	matches := []string{babyEndOfTextString, "foo " + babyEndOfTextString, babyEndOfTextString + " bar", "foo " + babyEndOfTextString + " bar", "foo " + babyEndOfTextString + " bar " + babyEndOfTextString}
	for _, match := range matches {
		match1 := bpe.specialTokenRegex.FindStringSubmatch(match)
		must(t, len(match1) == 2, "invalid bpe.specialTokenRegex match(%q) not found", match)
		must(t, match1[1] == babyEndOfTextString, "invalid bpe.specialTokenRegex match(%q) not correct", match)
	}
	nonMatches := []string{"", "foo", "<br/> ", "<|not_special|>"}
	for _, nonMatch := range nonMatches {
		must(t, bpe.specialTokenRegex.FindStringSubmatch(nonMatch) == nil, "invalid bpe.specialTokenRegex match(%q) WAS found", nonMatch)
	}
}

func TestBPETokenizer_EncodeDecode(t *testing.T) {
	// This test suite is fairly basic because BPETokenizer gets thoroughly
	// tested as part of the encoding packages' tests. There is one complicated
	// UTF-8 token in the baby set ("\xe2\x80"), so we include some unicode
	// characters that start with that, like •, ‹, ›. We do, however, test
	// the different special-token handling modes here.

	// bpe1 is a "baby" tokenizer initialized with special-tokens-as-special
	// enabled
	bpe1, err := getBabyBPETokenizer(false, []string{babyEndOfTextString})
	must(t, err == nil, "init bpe: %v", err)
	tests1 := []bpeTest{
		{"", []int{}},
		{"a", []int{64}},
		{"a a", []int{64, 257}},
		{"Write 3 knock-knock jokes.", []int{54, 81, 270, 68, 220, 18, 479, 77, 420, 74, 12, 74, 77, 420, 74, 474, 482, 274, 13}},
		{"(can't, you're, she'll)", []int{7, 66, 272, 470, 11, 345, 6, 260, 11, 264, 258, 6, 297, 8}},
		{"¿Qué hora es?", []int{126, 123, 48, 84, 127, 102, 289, 273, 64, 220, 274, 30}},
		{"• Bulleted \n• List\n", []int{447, 95, 347, 377, 293, 83, 276, 220, 198, 447, 95, 406, 396, 198}},
		{"‹ have this ›", []int{447, 117, 423, 428, 220, 447, 118}},
		{babyEndOfTextString, []int{babyEndOfTextToken}},
		{"For your information" + babyEndOfTextString, []int{37, 273, 345, 81, 287, 69, 273, 76, 341, babyEndOfTextToken}},
	}

	// none of these should error
	runBpeTests(t, bpe1, "tests1", tests1, false)

	// bpe2 is a baby tokenizer that declines to encode special tokens
	bpe2, err := getBabyBPETokenizer(false, []string{})
	must(t, err == nil, "init bpe: %v", err)
	tests2 := []bpeTest{
		{babyEndOfTextString, nil},
		{"For your information" + babyEndOfTextString, nil},
	}

	// all of these must error
	runBpeTests(t, bpe2, "tests2", tests2, true)

	// bpe3 encodes special tokens as text
	bpe3, err := getBabyBPETokenizer(true, []string{})
	must(t, err == nil, "init bpe: %v", err)
	tests3 := []bpeTest{
		{babyEndOfTextString, []int{27, 91, 437, 78, 69, 83, 68, 87, 83, 91, 29}},
		{"For your information" + babyEndOfTextString, []int{37, 273, 345, 81, 287, 69, 273, 76, 341, 27, 91, 437, 78, 69, 83, 68, 87, 83, 91, 29}},
	}

	// no errors for these
	runBpeTests(t, bpe3, "tests3", tests3, false)

	// decode error
	_, err = bpe3.Decode([]int{higherThanAnyToken})
	if err == nil {
		t.Errorf("BPETokenizer.Decode() did not error on invalid token")
	}
}

func runBpeTests(t *testing.T, bpe *BPETokenizer, suiteName string, tests []bpeTest, expectErrors bool) {
	t.Helper()
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s:encode(%q)", suiteName, tt.text), func(t *testing.T) {
			// test encoding
			got, err := bpe.Encode(tt.text)
			if err != nil && !expectErrors || err == nil && expectErrors {
				t.Errorf("BPETokenizer.Encode() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.tokens) {
				t.Errorf("BPETokenizer.Encode() = %#v, want %#v", got, tt.tokens)
			}

			// test count method (which actually calls Encode again)
			gotCount := bpe.Count(tt.text)
			if gotCount != len(tt.tokens) {
				t.Errorf("BPETokenizer.Count() = %d, want %d", gotCount, len(tt.tokens))
			}

			if expectErrors {
				// no need to test decode if encoding was supposed to fail
				return
			}

			// test decoding (should never error)
			got2, err := bpe.Decode(tt.tokens)
			if err != nil {
				t.Errorf("BPETokenizer.Decode() error = %v", err)
			}
			if got2 != tt.text {
				t.Errorf("BPETokenizer.Decode() = %q, want %q", got2, tt.text)
			}
		})
	}
}

func TestBPETokenizer_applyBPE(t *testing.T) {
	// This method is only used by Encode() and the tests above cover it 99%.
	// The only additional test case is for empty input.
	bpe, err := getBabyBPETokenizer(false, []string{})
	must(t, err == nil, "init bpe: %v", err)
	tokens := bpe.applyBPE([]byte{})
	if len(tokens) != 0 {
		t.Errorf("BPETokenizer.ApplyBPE([]byte{}) = %#v, want nil or empty slice", tokens)
	}
}
