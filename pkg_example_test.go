// This example demonstrates encoding and decoding a sample string using the
// cl100k_base tokenizer.
package gotoken_test

import (
	"fmt"
	"log"

	"github.com/peterheb/gotoken"
	_ "github.com/peterheb/gotoken/cl100kbase"
)

var tok gotoken.Tokenizer

func Example() {
	// Instantiate the tokenizer by name. The _ import above registers the
	// tokenizer with the encoding "cl100k_base". Consult your model's
	// documentation for information on which tokenizer to use with which model.
	tok, err := gotoken.GetTokenizer("cl100k_base")
	if err != nil {
		log.Fatal(err)
	}

	// Encode some text
	input := "Salutations, world! 😄"
	encoded, err := tok.Encode(input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("input: %#v\n", input)
	fmt.Printf("encoded: %#v\n", encoded)

	// Decode the encoded text
	decoded, err := tok.Decode(encoded)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("decoded: %#v\n", decoded)
}
