# Gotoken `bench` example

`bench` is a simple benchmarking tool for profiling gotoken.

## Requirements

This example app expects a 1GB test file that is not checked in to the
repository. To download it, use the `testdata/get_wiki_1gb.py` script.

## Usage

The benchmark is a CLI app. By default, all encodings are benchmarked in
single-threaded and multi-threaded mode. To limit the benchmark to a specific
encoding or number of threads, use the `-encoding` or `-threads` parameters,
like this:

- `./bench -encoding r50k_base`
- `./bench -threads 1`
- `./bench -encoding cl100k_base -threads 16`

Additionally, the `-pprof` flag can be used to write out CPU profiling data.
This will be saved to `./bench.pprof`, and can be accessed by running:

- `go tool pprof -http :8080 bench bench.pprof`
