// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package internal

// serializedTrie is a serialized trie that acts as a read-only, precomputed
// map[string]int.
type serializedTrie []uint32

// TrieLookup returns the index of the given input in a serialized trie. It
// returns -1 if the input is not present, or its token# otherwise. This is
// exported for use in gen.go.
func TrieLookup(trie []uint32, input []byte) int {
	return serializedTrie(trie).Lookup(input)
}

// Lookup returns the index of the given input in a serialized trie. It returns
// -1 if the input is not present, or its token# otherwise.
func (trie serializedTrie) Lookup(input []byte) int {
	if len(input) == 0 {
		return -1
	}
	pos := 0
	trieValue := 0
	for i, char := range input {
		trieValue = int(trie[pos])
		pos++
		childCount := trieValue & 0xff
		if childCount == 0 {
			// 0 means 256, which means a binary search is not necessary because
			// all 256 children are present in order.
			trieValue = int(trie[pos+int(char)])
		} else {
			// Binary search the children for the character we are looking for
			left := 0
			right := childCount
			for left < right {
				mid := left + (right-left)/2
				if byte(trie[pos+mid]&0xff) < char {
					left = mid + 1
				} else {
					right = mid
				}
			}

			if left == childCount {
				return -1
			}

			childIndex := pos + left
			childValue := byte(trie[childIndex] & 0xff)
			if childValue != char {
				return -1
			}
			// binary search successful
			trieValue = int(trie[childIndex])
		}

		// trieValue contains serialized trie node in this format:
		//
		// bits 31-9: token#, bit 8: 1 for leaf, bits 7-0: byte in input to match
		//    OR
		// bits 31-9: token#, bit 8: 0 for interior node, bits 7-0: byte in input

		// Is this a leaf? if so, this is the end of our search
		if trieValue&0x100 != 0 {
			if i == len(input)-1 {
				// If this was the last byte, we found the token! return token#
				return trieValue >> 9
			}
			// Otherwise, we ran out of trie before bytes: the token was not found
			return -1
		}

		// Otherwise, continue traversing the trie at the specified child node
		pos = trieValue >> 9
	}

	// If we made it to the end of the input, then trie[pos] points to a node
	// header. The high 20 bits contain (tokenIndex+1).
	return int(trie[pos]>>8) - 1
}
