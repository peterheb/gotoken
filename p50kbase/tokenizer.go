// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package p50kbase

import (
	"github.com/peterheb/gotoken"
	"github.com/peterheb/gotoken/internal"
)

// These special tokens are defined by this encoding.
const (
	EndOfText = "<|endoftext|>"
	FIMPrefix = "<|fim_prefix|>"
	FIMMiddle = "<|fim_middle|>"
	FIMSuffix = "<|fim_suffix|>"
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

// getTokenizerBase returns a BPE tokenizer that uses the OpenAI p50k_base
// encoding.
func getTokenizerBase(allowSpecialAsText bool, allowedSpecial []string) (gotoken.Tokenizer, error) {
	return internal.NewBPETokenizer(&internal.BPEParams{
		Name:           "p50k_base",
		Splitter:       internal.GPT2Splitter,
		ByteEncoder:    byteToToken,
		DecoderMap:     tokenList,
		EncoderTrie:    tokenTrie,
		SpecialTokens:  map[string]int{EndOfText: 50256},
		BytePairLookup: pairsToToken,
	}, allowSpecialAsText, allowedSpecial)
}

// getTokenizerEdit returns a BPE tokenizer that uses the OpenAI p50k_edit
// variation of p50k_base.
func getTokenizerEdit(allowSpecialAsText bool, allowedSpecial []string) (gotoken.Tokenizer, error) {
	return internal.NewBPETokenizer(&internal.BPEParams{
		Name:        "r50k_edit",
		Splitter:    internal.GPT2Splitter,
		ByteEncoder: byteToToken,
		DecoderMap:  tokenList,
		EncoderTrie: tokenTrie,
		SpecialTokens: map[string]int{
			EndOfText: 50256,
			FIMPrefix: 50281,
			FIMMiddle: 50282,
			FIMSuffix: 50283,
		},
		BytePairLookup: pairsToToken,
	}, allowSpecialAsText, allowedSpecial)
}

func init() {
	gotoken.RegisterTokenizer("p50k_base", getTokenizerBase)
	gotoken.RegisterTokenizer("p50k_edit", getTokenizerEdit)
}
