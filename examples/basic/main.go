// The basic example is an introduction to using gotoken.
package main

import (
	"fmt"
	"log"

	// The second import is for the data for r50k_base; substitute with one or
	// more encodings as needed!
	"github.com/peterheb/gotoken"
	_ "github.com/peterheb/gotoken/r50kbase"
)

func main() {
	// Instantiate our tokenizer with the "r50k_base" encoding
	tok, err := gotoken.GetTokenizer("r50k_base")
	if err != nil {
		log.Fatal(err)
	}

	// Encode and decode a string
	input := "Salutations, world! 😄"
	encoded, err := tok.Encode(input)
	if err != nil {
		log.Fatal(err)
	}
	decoded, err := tok.Decode(encoded)
	if err != nil {
		log.Fatal(err)
	}

	// Print the results. There is a tok.Count(input), but it actually just calls
	// Encode() and returns the length.
	fmt.Printf("token count:   %d\n", len(encoded))
	fmt.Printf("input string:  %#v\n", input)
	fmt.Printf("tokenized:     %#v\n", encoded)

	// Make strings out of every token
	tokenStr := make([]string, len(encoded))
	for i, t := range encoded {
		if tokenStr[i], err = tok.Decode([]int{t}); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("token values:  %#v\n", tokenStr)
	fmt.Printf("round-tripped: %#v\n", decoded)

	if input != decoded {
		log.Fatal("round-trip failed")
	}
}
