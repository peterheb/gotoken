// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package cl100kbase

import (
	"unicode"
	"unicode/utf8"
)

// replacementChar is the Unicode replacement character, used internally as a
// placeholder for broken UTF-8 code points
const replacementChar = rune(0xfffd)

// cl100KBaseSplitter is a SplitterFunc that implements the regex:
// `(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+`
func cl100KBaseSplitter(input []byte) [][]byte {
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
// match. Because of the construction of the regex, it always matches at least
// one character. Must be called with a non-empty input.
func getMatchLength(input []byte) int {
	cc := len(input)
	pos, next := 0, 0 // offset of current rune and next rune
	var c rune        // current rune

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
		// (?i:'s|'t|'re|'ve|'m|'ll|'d)
		switch input[1] {
		case 's', 't', 'm', 'd', 'S', 'T', 'M', 'D':
			return 2
		case 'r', 'v', 'R', 'V':
			if cc >= 3 && (input[2] == 'e' || input[2] == 'E') {
				return 3
			}
		case 'l', 'L':
			if cc >= 3 && (input[2] == 'l' || input[2] == 'L') {
				return 3
			}
		}
	}

	// [^\r\n\p{L}\p{N}]?\p{L}+ ... first [^\p{L}]? is elided as it simplifies away
	isLetter := unicode.IsLetter(c)
	isNumber := unicode.IsNumber(c)
	peek := replacementChar
	if next < cc {
		peek, _ = utf8.DecodeRune(input[next:])
	}
	if isLetter || (cc >= 2 && !isNumber && c != '\r' && c != '\n' && unicode.IsLetter(peek)) {
		for pos < cc {
			consume()
			if !unicode.IsLetter(c) {
				break
			}
		}
		return pos
	}

	// \p{N}{1,3}
	if isNumber {
		count := 0 // max 3 digits to match
		for pos < cc && count < 3 {
			consume()
			count++
			if !unicode.IsNumber(c) {
				break
			}
		}
		return pos
	}

	// match a space if it is present
	if c == ' ' && cc >= 2 {
		consume()
	}

	if !unicode.IsSpace(c) && !unicode.IsLetter(c) && !unicode.IsNumber(c) {
		// ` ?[^\s\p{L}\p{N}]+[\r\n]*` ... space already matched if present
		for pos < cc {
			consume()
			if unicode.IsSpace(c) || unicode.IsLetter(c) || unicode.IsNumber(c) {
				break
			}
		}
		for pos < cc && (c == '\r' || c == '\n') {
			consume()
		}
		return pos
	}

	// |\s+(?!\S)|\s+
	for pos < cc && unicode.IsSpace(c) {
		consume()
	}
	if pos >= 2 && pos < cc && !unicode.IsSpace(c) {
		// in a multi-space run, if there is a "next" non-space character, back
		// up and save the last space to match with that character
		return pos - 1
	}

	return pos
}
