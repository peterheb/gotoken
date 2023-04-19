// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package main

import (
	"sort"
)

// preallocatedSlots is the default capacity for children in new nodes.
const preallocatedSlots = 4

// TrieNode is a node in a trie data structure in memory.
type TrieNode struct {
	value    byte
	children []*TrieNode
	index    int
	fixup    int // used when serializing the trie
}

// BuildTrie builds a trie from the given tokens. The root of the constructed
// tree is returned.
func BuildTrie(tokens []string) *TrieNode {
	root := newTrieNode(0) // the value of the root node is ignored
	for index, token := range tokens {
		root.insert(token, index)
	}
	return root
}

// newTrieNode initializes a new TrieNode.
func newTrieNode(value byte) *TrieNode {
	return &TrieNode{
		value:    value,
		children: make([]*TrieNode, 0, preallocatedSlots),
		index:    -1,
	}
}

// insert inserts the given token into the trie. The index is the index of the
// token in the original list of tokens. The "value" saved in the trie in each
// node is the byte value of the character at that position in the token.
// Children are maintained in sorted order by value.
func (node *TrieNode) insert(token string, index int) {
	current := node
	for _, char := range []byte(token) {
		insertPos := sort.Search(len(current.children), func(i int) bool {
			return current.children[i].value >= char
		})

		if insertPos < len(current.children) && current.children[insertPos].value == char {
			current = current.children[insertPos]
			continue
		}

		newNode := newTrieNode(char)
		if insertPos == len(current.children) {
			current.children = append(current.children, newNode)
		} else {
			current.children = append(current.children[:insertPos+1], current.children[insertPos:]...)
			current.children[insertPos] = newNode
		}
		current = newNode
	}

	current.index = index
}

// lookup traverses a trie looking for "token" and returns the token index of
// the input string. If the input string is not found, -1 is returned.
func (node *TrieNode) lookup(token string) int {
	current := node
	for _, char := range []byte(token) {
		searchPos := sort.Search(len(current.children), func(i int) bool {
			return current.children[i].value >= char
		})

		if searchPos < len(current.children) && current.children[searchPos].value == char {
			current = current.children[searchPos]
		} else {
			return -1
		}
	}

	return current.index
}

// walk calls fn(node) and then recursively calls fn on all children. It returns
// the number of nodes visited (including itself). When called on the topmost
// node of a trie, the appropriate value to pass for depth is 0.
func (node *TrieNode) walk(fn func(*TrieNode, int), depth int) int {
	count := 1
	if fn != nil {
		fn(node, depth)
	}
	for _, child := range node.children {
		count += child.walk(fn, depth+1)
	}
	return count
}

// serialize serializes the trie into a flat array of uint32s.
//
//   - uint32: node header
//     -- bits 0-7: child count for this node
//     -- bits 8-31: token# of this node plus 1 (-1 encodes as 0, 0 as 1, etc.)
//   - []uint32: child data, one per child in sorted order
//     -- bits 0-7: byte value specifying which child this is
//     -- bit 8: 1 = this is a leaf , 0 = this points to a child node
//     -- bits 9-31: leaf: token# of child (not +1); node: ofs of child node
//
// This structure is designed to be easy to walk with the associated lookup()
// function. We output the trie one layer at a time, in the hopes of creating a
// little cache locality with the topmost layers that will get hit a lot.
// Ultimately, however, cache locality is not a trie's strength.
func (node *TrieNode) serialize() []uint32 {
	// Determine the size of the output and depth of the trie by walking it
	outputSize, maxDepth := 0, 0
	node.walk(func(node *TrieNode, depth int) {
		if len(node.children) > 0 {
			outputSize += len(node.children) + 1
		}
		maxDepth = max(maxDepth, depth)
	}, 0)

	// Allocate the output array
	output := make([]uint32, outputSize)
	i := 0

	// Walk the trie maxDepth+1 times, outputting each layer one by one. The
	// only important thing about the order from a correctness perspective is
	// that parent nodes must get emitted before child nodes, so that we can fix
	// up the parent's pointer to each child.
	for layer := 0; layer <= maxDepth; layer++ {
		node.walk(func(node *TrieNode, depth int) {
			if depth == layer {
				if node.fixup > 0 {
					// Modify our parent to point to where we landed in the output
					output[node.fixup] |= uint32((i & 0x7fffff) << 9)
				}
				if len(node.children) > 0 {
					// Leaf nodes do not get emitted; their info is stored
					// inside their parent's child pointer.
					output[i] = uint32(((node.index+1)&0xffffff)<<8 | (len(node.children) & 0xff))
					i++
					for _, child := range node.children {
						if len(child.children) == 0 {
							// Store leaf as value | index with bit 8 set
							output[i] = uint32(int(child.value) | child.index<<9 | 0x100)
						} else {
							// Store child pointer as value with bit 8 clear, pending fixup
							output[i] = uint32(int(child.value))
							child.fixup = i
						}
						i++
					}
				}
			}
		}, 0)
	}

	assert(len(output) == i, "serialized trie not the expected size: got %d, expected %d", i, len(output))
	return output
}
