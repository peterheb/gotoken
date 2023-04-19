// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package internal

import (
	"unicode"
	"unicode/utf8"
)

// replacementChar is the Unicode replacement character.
const replacementChar = 0xfffd

// GPT2Splitter implements the splitter function used by r50k_base and p50k_base
// to split text before byte-pair encoding. It is located in the internal
// package since it is shared between multiple encodings.
func GPT2Splitter(input []byte) [][]byte {
	pos := 0
	matches := make([][]byte, 0, len(input)/4)
	for pos < len(input) {
		matchLength := getMatchLength(input[pos:])
		matches = append(matches, input[pos:pos+matchLength])
		pos += matchLength
	}
	return matches
}

// getMatchLength runs a match against "input" and returns the length of the
// match, in bytes. Because of the construction of the regex, it always matches
// at least one byte. Must be called with a non-empty input.
func getMatchLength(input []byte) int {
	cc := len(input)
	pos, next := 0, 0 // offset of current rune and next rune
	var c rune        // current rune

	// The way getMatchLength works is:
	//
	// - c contains the "current" rune we have consumed and are evaluating
	// - pos is the offset of c in the input
	// - next is the offset of the rune after 'c'
	// - if input[pos] contains invalid utf8, we still consume 1 byte and pass it
	//   through. It is treated like replacementChar for classification purposes.

	// consume() advances c, pos, and next
	consume := func() {
		var size int
		pos = next
		c, size = utf8.DecodeRune(input[pos:])
		if c == utf8.RuneError {
			c = replacementChar
		}
		next = pos + size
	}

	consume() // first consume is just setup since pos==next!
	if c == '\'' && cc >= 2 {
		// 's|'t|'re|'ve|'m|'ll|'d
		switch input[1] {
		case 's', 't', 'm', 'd':
			return 2
		case 'r', 'v':
			if cc > 2 && (input[2] == 'e') {
				return 3
			}
		case 'l':
			if cc > 2 && (input[2] == 'l') {
				return 3
			}
		}
	}

	// match a space if it is present-- covers " ?" and "\s+" below
	if c == ' ' && cc >= 2 {
		consume()
	}

	if unicode.IsLetter(c) {
		// " ?\p{L}+"
		for pos < cc {
			consume()
			if !unicode.IsLetter(c) {
				break
			}
		}
		return pos
	} else if unicode.IsNumber(c) {
		// " ?\p{N}+"
		for pos < cc {
			consume()
			if !unicode.IsNumber(c) {
				break
			}
		}
		return pos
	} else if !unicode.IsSpace(c) {
		// " ?[^\s\p{L}\p{N}]+"
		for pos < cc {
			consume()
			if unicode.IsSpace(c) || unicode.IsLetter(c) || unicode.IsNumber(c) {
				break
			}
		}
		return pos
	}

	// "\s+(?!\S)|\s+"
	for pos < cc && unicode.IsSpace(c) {
		consume()
	}
	if pos >= 2 && pos < cc && !unicode.IsSpace(c) {
		// in a multi-space run where there is a "next" character, back up and
		// don't consume the last space, so it can match with that character
		return pos - 1
	}

	return pos
}
