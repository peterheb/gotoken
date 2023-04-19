#!/usr/bin/env python3
#
# This script generates the "ground truth" tokenization results from
# the tiktoken library for our test suite. Be sure to run this in a
# a venv with `pip install tiktoken` already installed.


import tiktoken
import json
import os


def encode_test_cases(encoding_name, input_file, output_file):
    encoding = tiktoken.get_encoding(encoding_name)

    print(f"{encoding_name}: writing to {output_file}... ", end="")
    lines, tokens = 0, 0
    with open(input_file, 'r') as f, open(output_file, 'w') as f_out:
        for line in f:
            encoded_line = encoding.encode(
                line.rstrip('\n'), allowed_special={""})
            serialized_line = json.dumps(encoded_line)
            f_out.write(serialized_line + "\n")
            lines += 1
            tokens += len(encoded_line)

    print(f"OK! ({lines} test cases, {tokens} tokens)")


if os.environ.get('GOTOKEN_TEST_1GB'):
    # large test suite-- not for most people!
    encode_test_cases("r50k_base",
                    "pae-enwiki-2023-04-1gb.txt", "r50k_1gb.txt")
    encode_test_cases("p50k_base",
                    "pae-enwiki-2023-04-1gb.txt", "p50k_1gb.txt")
    encode_test_cases("cl100k_base",
                    "pae-enwiki-2023-04-1gb.txt", "cl100k_1gb.txt")
else:
    # standard test suite
    encode_test_cases("r50k_base", "samples.txt", "r50k_base.txt")
    encode_test_cases("p50k_base", "samples.txt", "p50k_base.txt")
    encode_test_cases("cl100k_base", "samples.txt", "cl100k_base.txt")

