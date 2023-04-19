// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package internal

import (
	"testing"
)

const (
	testInput        = "../testdata/samples.txt"
	testExpected     = "../testdata/r50k_base.txt"
	testDoesNotExist = "../testdata/::does_not_exist::.txt"
	testNotJOSN      = "../testdata/gen_ground_truth.py"
	testInputLine    = "samples.txt#1"
)

func TestTestPairReader(t *testing.T) {
	tpr, err := NewTestPairReader(testInput, testExpected)
	if err != nil {
		t.Fatalf("opening test data: %v", err)
	}
	defer tpr.Close()

	// test file must have at least one test case
	tc, err := tpr.Next()
	if tc == nil || err != nil {
		t.Fatalf("%s: %v", tpr.CaseName(), err)
	}
	if tpr.Line() != 1 {
		t.Fatalf("tpr.Counter(): expected 1, got %d", tpr.Line())
	}
	if tpr.CaseName() != testInputLine {
		t.Fatalf("tpr.CaseName(): expected %s, got %s", testInputLine, tpr.CaseName())
	}

	// exhaust the test file(s)
	for {
		tc, err = tpr.Next()
		if tc == nil && err == nil {
			break
		}
		if err != nil {
			t.Fatalf("%s: %v", tpr.CaseName(), err)
		}
	}
	tc, err = tpr.Next()
	if tc != nil || err != nil {
		t.Fatalf("tpr.Next() did not return (nil, nil) after EOF")
	}
	tpr.Close() // ok to call multiple times

	_, err = NewTestPairReader(testDoesNotExist, testExpected)
	if err == nil {
		t.Fatalf("expected error loading non-existent test data")
	}

	_, err = NewTestPairReader(testInput, testDoesNotExist)
	if err == nil {
		t.Fatalf("expected error loading non-existent test data")
	}

	tpr2, err := NewTestPairReader(testInput, testNotJOSN)
	if err != nil {
		t.Fatalf("opening test data (not-JSON): %v", err)
	}
	defer tpr2.Close()
	for {
		tc, err := tpr2.Next()
		if tc == nil && err == nil {
			break
		} else if err == nil {
			t.Fatalf("expected error loading non-JSON test data")
		}
	}
}
