// The bench example is a synthetic benchmark that tokenizes every line in a
// test file.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/peterheb/gotoken"
	_ "github.com/peterheb/gotoken/cl100kbase"
	_ "github.com/peterheb/gotoken/p50kbase"
	_ "github.com/peterheb/gotoken/r50kbase"
)

// See: testdata/get_wiki_1gb.py to download the test data file.

func main() {
	// Parse flags
	threads := flag.Int("threads", 0, "Number of threads to use (0 = demo)")
	doProfile := flag.Bool("pprof", false, "Enable profiling")
	encoding := flag.String("encoding", "all", "Tokenizer encoding to use, default \"all\" (r50k_base, p50k_base, cl100k_base, all)")
	src := flag.String("src", "../../testdata/pae-enwiki-2023-04-1gb.txt", "Path to the test data file with one entry per line")
	flag.Parse()

	// Validate the specified encoding
	encodings := []string{"r50k_base", "p50k_base", "cl100k_base"}
	if *encoding != "all" {
		_, err := gotoken.GetTokenizer(*encoding)
		onErrFatalf(err, "unknown encoding: %q", *encoding)
		encodings = []string{*encoding}
	}

	// Validate the provided thread count
	if *threads > runtime.NumCPU()*4 {
		fmt.Printf("ignoring '-threads %d' (max=%d)\n", *threads, runtime.NumCPU()*4)
		*threads = 0
	}

	// Enable profiling if requested
	if *doProfile {
		f, err := os.Create("bench.pprof")
		onErrFatalf(err, "create bench.pprof")
		fmt.Println("outputting profiling data to bench.pprof")
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Run the benchmark for the specified encoding(s)
	for _, enc := range encodings {
		if *threads != 0 {
			// Run the benchmark with the specified number of threads
			runBenchmark(*src, enc, *threads)
		} else {
			// Run the benchmark with 1, 2, 4, 8, etc. up to NumCPU
			th := 1
			for th <= runtime.NumCPU() {
				runBenchmark(*src, enc, th)
				th *= 2
			}
		}
	}
}

func runBenchmark(src string, encoding string, threads int) {
	// Initialize encoder
	tok, err := gotoken.GetTokenizer(encoding)
	onErrFatalf(err, "create tokenizer")

	// Pre-load the file into RAM. This is a synthetic benchmark focusing on the
	// tokenizer, so we want to isolate the impact I/O has.
	data, err := os.ReadFile(src)
	onErrFatalf(err, "read %s", src)

	startTime := time.Now()
	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	i := 0
	if threads > 1 {
		// Parallel for pattern using goroutines and a semaphore channel
		sem := make(chan struct{}, threads)
		for scanner.Scan() {
			line := scanner.Bytes()
			i++
			sem <- struct{}{}
			go func(line string, i int) {
				defer func() { <-sem }()
				_, err := tok.Encode(line)
				onErrFatalf(err, "encode[line=%d] %s", i, line)
			}(string(line), i)
		}
		// Wait for final goroutines to finish
		for j := 0; j < cap(sem); j++ {
			sem <- struct{}{}
		}
	} else {
		// Single-threaded, just use a simpler loop!
		for scanner.Scan() {
			line := scanner.Text()
			i++
			_, err := tok.Encode(line)
			onErrFatalf(err, "encode[line=%d] %s", i, line)
		}
	}
	onErrFatalf(scanner.Err(), "bufio.Scanner: %s", src)
	dur := time.Since(startTime)
	durStr := fmt.Sprintf("%d:%02d.%02d", int(dur.Minutes()), int(dur.Seconds())%60, int(dur.Milliseconds()%1000)/10)
	fmt.Printf("%-13q (threads=%2d) elapsed time: %s sec, %.2f MiB/sec\n", encoding,
		threads, durStr, float64(len(data))/dur.Seconds()/1024/1024)
}

// onErrFatalf prints a message and ends the program if err!=nil.
func onErrFatalf(err error, format string, args ...any) {
	if err != nil {
		fmt.Printf(format, args...)
		fmt.Printf(": %v\n", err)
		os.Exit(1)
	}
}
