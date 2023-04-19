// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package cl100kbase

import (
	"reflect"
	"testing"
)

func TestCL100KBaseSplitter(t *testing.T) {
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
			name: "text",
			args: "This is a regular sentence without contractions.",
			want: []string{"This", " is", " a", " regular", " sentence", " without", " contractions", "."},
		},
		{
			name: "text and numbers",
			args: "a a 1234567890 z",
			want: []string{"a", " a", " ", "123", "456", "789", "0", " z"},
		},
		{
			name: "text with two digit number",
			args: "The quick brown fox jumps over 13 lazy dogs.",
			want: []string{"The", " quick", " brown", " fox", " jumps", " over", " ", "13", " lazy", " dogs", "."},
		},
		{
			name: "contractions",
			args: "I'm a test case, aren't I? I'd like to know if you'll be able to tokenize me correctly.",
			want: []string{"I", "'m", " a", " test", " case", ",", " aren", "'t", " I", "?", " I", "'d", " like", " to", " know", " if", " you", "'ll", " be", " able", " to", " tokenize", " me", " correctly", "."},
		},
		{
			name: "numbers, and newline after period",
			args: "I have 3 apples and 4 oranges.\n",
			want: []string{"I", " have", " ", "3", " apples", " and", " ", "4", " oranges", ".\n"},
		},
		{
			name: "large run of quoted spaces",
			args: `These spaces "                         " are one token in p50k.`,
			want: []string{"These", " spaces", " \"", "                        ", " \"", " are", " one", " token", " in", " p", "50", "k", "."},
		},
		{
			name: "other ASCII characters",
			args: "Hello! @username, did you check the #hashtag?",
			want: []string{"Hello", "!", " @", "username", ",", " did", " you", " check", " the", " #", "hashtag", "?"},
		},
		{
			name: "URL and punctuation sequences",
			args: "Test cases for https://github.com/peterheb/gotoken /***** ¯\\_(ツ)_/¯ ******/",
			want: []string{"Test", " cases", " for", " https", "://", "github", ".com", "/peterheb", "/gotoken", " /*****", " \u00af\\_(", "\u30c4", ")_/\u00af", " ******/"},
		},
		{
			name: "multi-byte unicode characters",
			args: "こんにちは、世界！",
			want: []string{"\u3053\u3093\u306b\u3061\u306f", "\u3001\u4e16\u754c", "\uff01"},
		},
		{
			name: "mixed scripts",
			args: "I'm learning العَرَبِيَّة and हिन्दी languages.",
			want: []string{"I", "'m", " learning", " \u0627\u0644\u0639", "\u064e\u0631", "\u064e\u0628", "\u0650\u064a", "\u064e\u0651", "\u0629", " and", " \u0939", "\u093f\u0928", "\u094d\u0926", "\u0940", " languages", "."},
		},
		{
			name: "emoji",
			args: "I'm happy 😃 and excited 🎉!",
			want: []string{"I", "'m", " happy", " 😃", " and", " excited", " 🎉!"},
		},
		{
			name: "emoji at end of string",
			args: "Hello, World! How are you today? 🌍",
			want: []string{"Hello", ",", " World", "!", " How", " are", " you", " today", "?", " 🌍"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cl100KBaseSplitter([]byte(tt.args)); !reflect.DeepEqual(asStrings(got), tt.want) {
				t.Errorf("CL100KBaseMatches() = %#v, want %#v", asStrings(got), tt.want)
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
