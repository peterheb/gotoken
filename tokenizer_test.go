// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package gotoken

import (
	"testing"
)

func TestMain(m *testing.M) {
	// For these tests, set up a mock tokenizer named "runes"
	RegisterTokenizer("runes", func(allowSAT bool, allowedSpc []string) (Tokenizer, error) {
		return &runeTokenizer{
			allowSpecialAsText:   allowSAT,
			allowedSpecialTokens: allowedSpc,
		}, nil
	})

	m.Run()
}

func TestGetTokenizer(t *testing.T) {
	// test basic gets
	tok, err := GetTokenizer("runes")
	if err != nil {
		t.Fatalf("GetTokenizer('runes'): %v", err)
	}
	if _, ok := tok.(*runeTokenizer); !ok {
		t.Fatalf("GetTokenizer('runes'): expected *runeTokenizer, got %T", tok)
	}

	_, err = GetTokenizer("does_not_exist")
	if err == nil {
		t.Fatalf("GetTokenizer('does_not_exist'): expected error, got nil")
	}

	// test TokenizerOpts
	tok1, err := GetTokenizer("runes", WithSpecialTokensAsText())
	if err != nil {
		t.Fatalf("GetTokenizer('runes', WithSpecialTokensAsText()): %v", err)
	}
	if tok1.(*runeTokenizer).allowSpecialAsText != true {
		t.Fatalf("GetTokenizer('runes', WithSpecialTokensAsText()) did not set AllowSpecialAsText")
	}

	tok2, err := GetTokenizer("runes", WithSpecialTokens("<|foo|>"))
	if err != nil {
		t.Fatalf("GetTokenizer('runes', WithSpecialTokens()): %v", err)
	}
	hasFoo := false
	for _, tok := range tok2.(*runeTokenizer).allowedSpecialTokens {
		if tok == "<|foo|>" {
			hasFoo = true
		}
	}
	if !hasFoo {
		t.Fatalf("GetTokenizer('runes', WithSpecialTokens()) did not set AllowedSpecialTokens")
	}
}

func TestListTokenizers(t *testing.T) {
	// In theory, we can't import our real tokenizers here, because it creates
	// an import cycle. In practice, "cl100k_base" is actually registered when
	// this test runs 🧐. I think it is leaking in from pkg_example_test.go.
	list := ListTokenizers()
	for _, name := range list {
		if name == "runes" {
			return // success
		}
	}

	t.Fatal("ListTokenizers() did not include 'runes'")
}

// runeTokenizer is a mock tokenizer that just returns runes as tokens. It
// ignores all tokenizer options.
type runeTokenizer struct {
	allowSpecialAsText   bool
	allowedSpecialTokens []string
}

func (at *runeTokenizer) Encode(s string) ([]int, error) {
	tokens := make([]int, 0, len(s))
	for _, c := range s {
		tokens = append(tokens, int(c))
	}
	return tokens, nil
}

func (at *runeTokenizer) Decode(tokens []int) (string, error) {
	runes := make([]rune, 0, len(tokens))
	for _, t := range tokens {
		runes = append(runes, rune(t))
	}
	return string(runes), nil
}

func (at *runeTokenizer) Count(s string) int {
	return len([]rune(s))
}

func (at *runeTokenizer) Allowed(s string) error {
	return nil
}
