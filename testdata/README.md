# Gotoken Testing Info

This directory contains test data for gotoken.

## Basic Test Set

The `samples.txt` file contains a small number of tests that provide basic
coverage of the tokenizer across different Unicode scripts and various cases
known to differ between encodings. Each line in the text file represents one
test case. Note that any "comments" at the top of the file are treated as a test
case starting with `"#"`, not as actual comments.

The Python script `gen_ground_truth.py` generates the "ground truth" tokenized
output for each test case. Ground truth files, named `{encoding}_base.txt`, are
included in the repository and contain JSON arrays of the expected tokens for
each test case.

## Wikipedia Partial Article Extract (1GB)

The `pae-enwiki-2013-04-1gb.txt` file is a 1GB extract from English Wikipedia,
consisting of over 700,000 lines, each under 4KiB. Because of its large size and
limited utility to most users of this library, this file is not checked in to
the repository. To download it, use the `get_wiki_1gb.py` script.

The `gen_ground_truth.py` script can also generate ground truth for this file.
To do so, run:

- `GOTOKEN_TEST_1GB=1 python3 gen_ground_truth.py`

The generated ground truth files are saved as `{encoding}_1gb.txt`.

To run the large test suite, use:

- `GOTOKEN_TEST_1GB=1 go test ./... -timeout 3600s`

## Fuzz Testing

The three tokenizers support Go's built in fuzz-testing framework. To run the
fuzz tests, use:

- `go test -fuzz=FuzzR50K github.com/peterheb.gotoken/r50kbase`
- `go test -fuzz=FuzzP50K github.com/peterheb.gotoken/p50kbase`
- `go test -fuzz=FuzzCL100K github.com/peterheb.gotoken/cl100kbase`

The base fuzzing corpus is built-in to the tests.
