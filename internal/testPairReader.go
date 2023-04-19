// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TestPair contains the input and expected output for a tokenization test case.
type TestPair struct {
	Input    string
	Expected []int
}

// TestPairReader reads a pair of text files that contain tokenization test cases.
// The rows from the file are read one-by-one.
type TestPairReader struct {
	inputScanner    *bufio.Scanner
	expectedScanner *bufio.Scanner
	inputF          *os.File
	expectedF       *os.File
	nameBase        string
	line            int
	isClosed        bool
}

// NewTestPairReader returns a new TestPairReader that reads from the given
// input and expected files. The parameter inputFile must point to a text file
// that contains one test case per line, and expectedFile should contain one
// JSON-encoded array of integers per line.
func NewTestPairReader(inputFile, expectedFile string) (*TestPairReader, error) {
	inputF, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	expectedF, err := os.Open(expectedFile)
	if err != nil {
		return nil, err
	}

	inputScanner := bufio.NewScanner(inputF)
	expectedScanner := bufio.NewScanner(expectedF)

	return &TestPairReader{
		inputScanner:    inputScanner,
		expectedScanner: expectedScanner,
		inputF:          inputF,
		expectedF:       expectedF,
		nameBase:        filepath.Base(inputFile),
	}, nil
}

// Next returns the next test case from the input and expected files. If the end
// of the files has been reached, Next() returns (nil, nil) indefinitely.
func (tpr *TestPairReader) Next() (*TestPair, error) {
	if tpr.isClosed {
		return nil, nil
	}

	// Read a line from both scanners. If one of the files ends, then we are
	// done.
	if tpr.inputScanner.Scan() && tpr.expectedScanner.Scan() {
		inputLine := tpr.inputScanner.Bytes()
		expectedLine := tpr.expectedScanner.Bytes()
		tpr.line++

		var expectedData []int
		err := json.Unmarshal(expectedLine, &expectedData)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", tpr.CaseName(), err)
		}

		return &TestPair{
			Input:    string(inputLine),
			Expected: expectedData,
		}, nil
	}

	// If we get here, one of the scanners reached EOF, so close the files and
	// signal the caller.
	tpr.Close()
	return nil, nil
}

// Line returns the the 1-based line number of the last lines read from the
// input and expected files.
func (tpr *TestPairReader) Line() int {
	return tpr.line
}

// CaseName returns "{inputFilename}#{line}" to identify the current test case.
func (tpr *TestPairReader) CaseName() string {
	return fmt.Sprintf("%s#%d", tpr.nameBase, tpr.line)
}

// Close closes the input and expected files. It is safe to call even on a
// closed reader.
func (fp *TestPairReader) Close() {
	if !fp.isClosed {
		fp.inputF.Close()
		fp.expectedF.Close()
		fp.isClosed = true
	}
}
