// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package internal

import (
	"reflect"
	"testing"
)

func TestGPT2Splitter(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args string
		want []string
	}{
		{
			name: "empty string",
			args: "",
			want: []string{},
		},
		{
			name: "basic example",
			args: "a a",
			want: []string{"a", " a"},
		},
		{
			name: "regular text",
			args: "This is a regular sentence without high-maintenance contractions.",
			want: []string{"This", " is", " a", " regular", " sentence", " without", " high", "-", "maintenance", " contractions", "."},
		},
		{
			name: "contractions",
			args: "I'm a test case, aren't I? I'd like to know if YOU'LL be able to tokenize me correctly.",
			want: []string{"I", "'m", " a", " test", " case", ",", " aren", "'t", " I", "?", " I", "'d", " like", " to", " know", " if", " YOU", "'", "LL", " be", " able", " to", " tokenize", " me", " correctly", "."},
		},
		{
			name: "numbers",
			args: "I have 6 apples and 8 oranges,\r\n    ...or 14 pieces of fruit.\n",
			want: []string{"I", " have", " 6", " apples", " and", " 8", " oranges", ",", "\r\n   ", " ...", "or", " 14", " pieces", " of", " fruit", ".", "\n"},
		},
		{
			name: "other ASCII characters",
			args: "Hello! @username, did you check the #hashtag?",
			want: []string{"Hello", "!", " @", "username", ",", " did", " you", " check", " the", " #", "hashtag", "?"},
		},
		{
			name: "large run of quoted spaces",
			args: `These spaces "                         " are one token in p50k.`,
			want: []string{"These", " spaces", " \"", "                        ", " \"", " are", " one", " token", " in", " p", "50", "k", "."},
		},
		{
			name: "multi-byte unicode characters",
			args: "こんにちは、世界！",
			want: []string{"こんにちは", "、", "世界", "！"},
		},
		{
			name: "mixed scripts (RTL)",
			args: "I'm learning العَرَبِيَّة and हिन्दी languages.",
			want: []string{"I", "'m", " learning", " \u0627\u0644\u0639", "\u064e", "\u0631", "\u064e", "\u0628", "\u0650", "\u064a", "\u064e\u0651", "\u0629", " and", " \u0939", "\u093f", "\u0928", "\u094d", "\u0926", "\u0940", " languages", "."},
		},
		{
			name: "emoji",
			args: "I'm happy 😃 and you're excited 🎉 she'll play!",
			want: []string{"I", "'m", " happy", " 😃", " and", " you", "'re", " excited", " 🎉", " she", "'ll", " play", "!"},
		},
		{
			name: "emoji at end of string",
			args: "Hello, World! How are you today? 🌍",
			want: []string{"Hello", ",", " World", "!", " How", " are", " you", " today", "?", " 🌍"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GPT2Splitter([]byte(tt.args)); !reflect.DeepEqual(asStrings(got), tt.want) {
				t.Errorf("GPT2Splitter() got %#v, want %#v", got, tt.want)
			}
		})
	}
}

// asStrings converts a slice of byte slices to a slice of strings.
func asStrings(parts [][]byte) []string {
	strings := make([]string, len(parts))
	for i, part := range parts {
		strings[i] = string(part)
	}
	return strings
}
