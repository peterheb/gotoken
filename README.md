# Gotoken <!-- omit from toc -->

[![PkgGoDev](https://pkg.go.dev/badge/github.com/peterheb/gotoken)](https://pkg.go.dev/github.com/peterheb/gotoken)
[![license: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Gotoken is a pure-Go implementation of the Python library
[openai/tiktoken](https://github.com/openai/tiktoken).

With gotoken, you can:

- Count tokens for billing or to limit request size.
- Perform specialized tokenization for OpenAI API calls.
- Better understand byte-pair encoding and its implementation.

## Table of Contents <!-- omit from toc -->

- [Installation](#installation)
- [Usage](#usage)
  - [Which encoding do I use?](#which-encoding-do-i-use)
  - [Dealing with special tokens](#dealing-with-special-tokens)
- [Differences from tiktoken](#differences-from-tiktoken)
- [Performance](#performance)
- [Version History](#version-history)
- [Acknowledgements](#acknowledgements)
- [License](#license)

## Installation

To add gotoken to a project, add it as a dependency with `go get`:

```bash
go get -u -v github.com/peterheb/gotoken
```

Gotoken uses Go modules, and requires Go 1.18 or later. It currently has no
external dependencies outside the standard library.

## Usage

To use gotoken, follow these steps:

- Pass an encoding name to `gotoken.GetTokenizer()` to receive a tokenizer
  instance.
- Optionally, use an option like `gotoken.WithSpecialTokens()` when creating a
  tokenizer to enable the encoding of special tokens, if required for your
  application.
- Use the `Encode`, `Decode`, and `Count` methods on the returned tokenizer.

Example from [examples/basic/main.go](examples/basic/main.go):

```go
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
```

The output of the example program is:

```text
token count:   7
input string:  "Salutations, world! 😄"
tokenized:     []int{17691, 83241, 11, 1917, 0, 27623, 226}
token values:  []string{"Sal", "utations", ",", " world", "!", " \xf0\x9f\x98", "\x84"}
round-tripped: "Salutations, world! 😄"
```

Some notes about this output:

- A token corresponds to a sequence of bytes, which can be a word, a piece of a
  word, white space, or some combination. Common words are often one token,
  while uncommon words are often split into multiple tokens. Word tokens often
  start with a space, like `" world"`.
- Tokens can also be partial Unicode code points, especially for Emoji or
  non-English scripts. For example, the emoji in the example above is `U+1F604`,
  which has a four-byte UTF-8 encoding, `"\xf0\x9f\x98\x84"`. These four bytes
  get split by the tokenizer across two tokens. Neither token makes sense on its
  own, but concatenated, they form the valid UTF-8 sequence for `😄`.

> **Keep in mind:** Slicing an `[]int` of tokens may cause the underlying string
> to be sliced between Unicode code points. This can lead to lost characters,
> garbled text, or replacement characters (`�`) appearing in a string if these
> token slices are later decoded. When tokenizing long text, it's recommended to
> split the text first at known-safe boundaries, and then tokenize those parts.
> Splitting a returned `[]int` of tokens may have unexpected results.

### Which encoding do I use?

The universe of OpenAI's LLMs is expanding rapidly, and there are many different
models. The gotoken library does not attempt to provide a mapping of models to
tokenizers; refer to OpenAI's documentation for this. However, as a general
guide, as of April 2023, the current models use `cl100k_base`, the previous
generation uses `p50k_base` or `p50k_edit`, and the oldest models use
`r50k_base`.

Gotoken focuses on OpenAI models and does not include tokenizers for other
models, such as BERT or LLaMa. However, the `r50k_base` tokenizer is compatible
with models that use GPT-2-compatible tokenization.

### Dealing with special tokens

Special tokens are strings that tokenize to unique token values outside the
regular range of byte-pair encoded tokens, like `"<|endoftext|>"`. Gotoken
mirrors the design of tiktoken and *disallows* all special tokens in the input
to `Encode()` by default. For example, attempting to tokenize this README file
with a default gotoken Tokenizer would fail with a wrapped `ErrSpecialToken`. A
comment in the tiktoken source explains:

> Special tokens are artificial tokens used to unlock capabilities from a model,
> such as fill-in-the-middle. We want to be careful about accidentally encoding
> special tokens since they can trick a model into doing something we don't want
> it to do.

Whether this presents a security issue in your application depends on how you
are using gotoken. Generally, a model should not treat special tokens encoded as
text any differently from other words in a prompt. To allow them to be encoded
this way, use the `WithSpecialTokensAsText()` option when creating a tokenizer:

```go
ttok, err := gotoken.GetTokenizer("cl100k_base", gotoken.WithSpecialTokensAsText())
```

With this option, the `cl100k_base` encoding would tokenize `"<|endoftext|>"` as
`{"<", "|", "endo", "ft", "ext", "|", ">"}`. The exact encoding will vary
depending on the tokenizer used. This should generally be safe right before
making an API call, or if just counting tokens.

To allow individual special tokens to be encoded with their *special values* and
be interpreted by the model, use the `WithSpecialTokens()` option, specifying a
list of allowed tokens by their string representations:

```go
stok, err := gotoken.GetTokenizer("cl100k_base", gotoken.WithSpecialTokens(cl100kbase.EndOfText))
```

The above tokenizer will encode `"<|endoftext|>"` with its special token value
in this encoding, `100257`. When using `Encode()` this way, ensure that any text
from external users has been sanitized to avoid unexpected behavior.

To control special token usage, it is valid to specify either option or both.
The possible behaviors are summarized in this table:

| Options Specified | Behavior |
|-------------------|-------------------|
| default (no options) | Return an error if a special token is encountered in the input. |
| only `WithSpecialTokensAsText()` | Encode all special tokens in the input as text. |
| only `WithSpecialTokens()` | Encode the specified special tokens with their true token values. Return an error if any other special token is encountered in the input. |
| both `WithSpecialTokensAsText()` and `WithSpecialTokens()` | Encode the specified special tokens with their true token values. Encode any other special tokens in the input as text. |

## Differences from tiktoken

Gotoken aims to produce identical outputs to the Python tiktoken library.

There are some differences in behavior for invalid UTF-8 sequences, due to the
intrinsic differences between Go strings and Python strings. Go strings are
UTF-8 `[]byte` sequences, while Python strings behave more like a Go `[]rune`
and are generally comprised of whole Unicode characters.

For instance, consider the string `"\xc0"`:

- In Python, `"\xc0"` is equivalent to `"À"`, Unicode code-point `U+00C0`,
  "Latin Capital Letter A with Grave".
- In Go, `"\xc0"` is a one-byte string that does not represent a Unicode code
  point. The string `"À"` is equal to the two-byte sequence `"\xc3\x80"`;
  `len("À") == 2` in Go. The Python equivalent to the invalid Go string would be
  `b"\xc0"`.

For invalid UTF-8 sequences, gotoken's `Encode()` returns a slice of tokens that
will successfully round-trip the invalid byte sequence, ensuring that `s ==
tok.Decode(tok.Encode(s))`. Tiktoken doesn't `encode()` UTF-8 strings directly.

Ultimately, this behavior difference shouldn't matter much in real-life usage,
since it only relates to what happens with invalid inputs.

> **A note about Unicode versions:** Go, as of version 1.20, supports Unicode
> 13.0, which is slightly out-of-date. Newly added Unicode code points do not
> have up-to-date metadata in the Go `unicode` library. This may result in
> gotoken returning different (but equivalent) tokenizations for inputs that
> contain these code points. That said, the model probably doesn't know what
> those code points mean, either.

## Performance

Gotoken employs precomputed data tables for encoding and decoding. These are
created with `go generate` and are compiled-in to the library. This approach
increases the size of compiled binaries by a few MB, but eliminates the need for
downloads or locally-cached data files during initialization.

Tokenizer instances are thread-safe. The benchmark
[examples/bench/main.go](examples/bench/main.go) measures performance by
tokenizing the lines of a 1GB test file. Here is an example run on a Ryzen
5700X CPU:

```text
$ ./bench -encoding cl100k_base
"cl100k_base" (threads= 1) elapsed time: 0:40.34 sec, 25.38 MiB/sec
"cl100k_base" (threads= 2) elapsed time: 0:22.80 sec, 44.90 MiB/sec
"cl100k_base" (threads= 4) elapsed time: 0:11.77 sec, 86.94 MiB/sec
"cl100k_base" (threads= 8) elapsed time: 0:06.56 sec, 155.90 MiB/sec
"cl100k_base" (threads=16) elapsed time: 0:04.43 sec, 230.68 MiB/sec
```

## Version History

- **v0.9.0** (2023-04-17)
  - Initial pre-release version.

## Acknowledgements

- [tiktoken](https://github.com/openai/tiktoken) is the official OpenAI
  open-source tokenizer.
- [SharpToken](https://github.com/dmitry-brazhenko/SharpToken) is an independent
  C# tokenizer implementation. Most of gotoken's standard test cases are
  borrowed from SharpToken. Thanks, @dmitry-brazhenko!

## License

Gotoken is licensed under the MIT License. Accordingly it is free to use,
re-mix, or adapt, in both commercial or non-commercial settings, per the terms
of the license. See [LICENSE](LICENSE) for the full license text.

This project and its author(s) are not affiliated with OpenAI.
