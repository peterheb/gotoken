// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package cl100kbase

import (
	"github.com/peterheb/gotoken"
	"github.com/peterheb/gotoken/internal"
)

// These special tokens are defined by this encoding.
const (
	EndOfText   = "<|endoftext|>"
	FIMPrefix   = "<|fim_prefix|>"
	FIMMiddle   = "<|fim_middle|>"
	FIMSuffix   = "<|fim_suffix|>"
	IMStart     = "<|im_start|>" // these are documented in the tiktoken README
	IMEnd       = "<|im_end|>"   // but aren't in the Python code
	EndOfPrompt = "<|endofprompt|>"
)

// pairsToToken is a lookup table that maps pairs of bytes to their token, or -1
// if the pair is not present in the encoding. This is used to bootstrap
// byte-pair-encoding and is generated from bytePairLookup in data.go.
var pairsToToken []int

func init() {
	// inflate bytePairLookup into pairsToToken
	pairsToToken = make([]int, 65536)
	for i := 0; i < 65536; i++ {
		pairsToToken[i] = -1
	}
	for _, pair := range bytePairLookup {
		pairsToToken[pair>>20] = int(pair) & 0xfffff
	}
}

// getTokenizer returns a BPE tokenizer that uses the OpenAI cl100k_base
// encoding.
func getTokenizer(allowSpecialAsText bool, allowedSpecial []string) (gotoken.Tokenizer, error) {
	return internal.NewBPETokenizer(&internal.BPEParams{
		Name:        "cl100k_base",
		Splitter:    cl100KBaseSplitter,
		ByteEncoder: byteToToken,
		DecoderMap:  tokenList,
		EncoderTrie: tokenTrie,
		SpecialTokens: map[string]int{
			EndOfText:   100257,
			FIMPrefix:   100258,
			FIMMiddle:   100259,
			FIMSuffix:   100260,
			IMStart:     100264,
			IMEnd:       100265,
			EndOfPrompt: 100276,
		},
		BytePairLookup: pairsToToken,
	}, allowSpecialAsText, allowedSpecial)
}

func init() {
	gotoken.RegisterTokenizer("cl100k_base", getTokenizer)
}
