# gotoken generator

This directory contains the generator that converts OpenAI `.tiktoken` files
into data embedded in Go source files.

## Usage

From this directory, run `go generate`. It will output the generated Go source
at `../{encoding}/data.go`, where `{encoding}` gets replaced with each of the
supported tokenizers.
