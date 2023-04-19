// Copyright 2023 Peter Hebert. Licensed under the MIT license.

// Package gotoken provides an OpenAI-compatible tokenization library similar to
// tiktoken. Its primary export is the Tokenizer interface, featuring Encode and
// Decode methods for converting strings to/from []int.
//
// Tokenizer encodings, such as r50kbase or cl100kbase, are available in
// separate packages that implement the Tokenizer interface. This design mirrors
// the image/png and image/jpeg packages' integration with the standard image
// library. Encoding packages self-register with gotoken.
//
// Encoding packages include built-in token dictionaries, which removes the need
// for external downloads or local file caches. However, these packages are
// relatively large (a few MB) and should only be imported when needed. At least
// one encoding package must be imported for gotoken to be able to tokenize
// text.
//
// Example of importing gotoken and a tokenizer encoding:
//
//	import (
//	    "github.com/peterheb/gotoken"
//	    _ "github.com/peterheb/gotoken/cl100kbase"
//	)
//
// The _ indicates that cl100kbase should be imported even without a direct
// reference in your code. Encoding packages have no public functions or types,
// but they do contain public constants defining special tokens.
//
// [tiktoken]: https://github.com/openai/tiktoken
package gotoken

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Tokenizer is the primary public interface provided by gotoken. It is
// implemented by encoding packages, like
// [github.com/peterheb/gotoken/r50kbase]. A Tokenizer is created using
// [GetTokenizer].
//
// Tokenizer supports four methods:
//
//   - Count returns the number of tokens in an input string, or 0 on error.
//   - Encode tokenizes an input string to an []int.
//   - Decode un-tokenizes an []int back to its string representation.
//   - Allowed returns an error if the input string contains any sequences
//     corresponding to special tokens that are not allowed by this tokenizer.
type Tokenizer interface {
	Count(input string) int
	Encode(input string) ([]int, error)
	Decode(input []int) (string, error)
	Allowed(input string) error
}

// Option is a functional option for a tokenizer, such as [WithSpecialTokens] or
// [WithSpecialTokensAsText].
type Option func(*tokenizerOptions)

// tokenizerOptions collects data from our functional options.
type tokenizerOptions struct {
	AllowSpecialAsText   bool
	AllowedSpecialTokens []string
}

// These errors can be returned by functions in this library. Errors will be
// wrapped with fmt.Errorf; use [errors.Is] or [errors.As] to check for the
// underlying error type.
var (
	ErrUnknownEncoding = errors.New("unknown tokenizer encoding")
	ErrInvalidToken    = errors.New("invalid token")
	ErrSpecialToken    = errors.New("unexpected special token found")
)

var (
	registered = make(map[string]func(bool, []string) (Tokenizer, error))
	regMu      sync.RWMutex
)

// GetTokenizer returns a tokenizer by its encoding name. If no matching
// registered encoding is found, an error is returned that wraps
// [ErrUnknownEncoding].
//
// GetTokenizer supports functional options to configure the returned Tokenizer.
// The default configuration, if no options are specified, disallows special
// tokens in the input.
//
// If special tokens are not applicable, using [WithSpecialTokensAsText] will
// allow the tokenizer to process any input string without raising an error. If
// special tokens should be supported by the Tokenizer, list the specific ones
// to allow using the option [WithSpecialTokens].
//
// The following encoding names are supported:
//
//   - "cl100k_base" in [github.com/peterheb/gotoken/cl100kbase]
//   - "p50k_base" and "p50k_edit" in [github.com/peterheb/gotoken/p50kbase]
//   - "r50k_base" in [github.com/peterheb/gotoken/r50kbase]
func GetTokenizer(encodingName string, opts ...Option) (Tokenizer, error) {
	regMu.RLock()
	defer regMu.RUnlock()

	// If options are provided, apply them.
	options := tokenizerOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	// Return a new tokenizer instance
	if tokenFactory, ok := registered[encodingName]; ok {
		return tokenFactory(options.AllowSpecialAsText, options.AllowedSpecialTokens)
	}

	return nil, fmt.Errorf("%w: %s", ErrUnknownEncoding, encodingName)
}

// ListTokenizers returns a list of all registered tokenizer encodings. These
// are the valid inputs to [GetTokenizer].
func ListTokenizers() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	encodings := make([]string, 0, len(registered))
	for encoding := range registered {
		encodings = append(encodings, encoding)
	}
	sort.Strings(encodings)
	return encodings
}

// RegisterTokenizer registers a tokenizer with the given name. This is
// typically called by the init function of a specific tokenizer's package.
func RegisterTokenizer(name string, tokFactory func(bool, []string) (Tokenizer, error)) {
	regMu.Lock()
	defer regMu.Unlock()
	registered[name] = tokFactory
}

// WithSpecialTokensAsText is a functional option for [GetTokenizer] that
// configures the tokenizer to treat special tokens as text. This allows strings
// like "<|endoftext|>" to be encoded as text tokens, rather than causing an
// encoding error (which is the default behavior).
func WithSpecialTokensAsText() func(*tokenizerOptions) {
	return func(opts *tokenizerOptions) {
		opts.AllowSpecialAsText = true
	}
}

// WithSpecialTokens is a functional option for [GetTokenizer] that configures
// the tokenizer to encode special tokens to their special token values. This
// should only be used when a Tokenizer is encoding trusted input.
func WithSpecialTokens(tokens ...string) func(*tokenizerOptions) {
	return func(opts *tokenizerOptions) {
		for _, tok := range tokens {
			opts.AllowedSpecialTokens = append(opts.AllowedSpecialTokens, tok)
		}
	}
}
