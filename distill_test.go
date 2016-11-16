package main

import (
	"testing"

	"github.com/pladdy/lumberjack"
)

func TestDrainChannel(t *testing.T) {
	lumberjack.Hush()

	// load up a channel
	c := make(chan error, 11)
	for i := 0; i < 11; i++ {
		c <- nil
	}

	// drain most of it
	drainChannel(c, 10)
	lastValue := <-c

	if lastValue != nil {
		t.Error("Expected:", nil, "Got:", lastValue)
	}
}

func TestLastString(t *testing.T) {
	type stringTests struct {
		input    []string
		expected string
	}

	var tests = []stringTests{
		{[]string{"one", "two", "three"}, "three"},
		{[]string{"this", "is", "last"}, "last"},
	}

	for _, test := range tests {
		if lastString(test.input) != test.expected {
			t.Error(
				"For", test.input,
				"expected", test.expected,
				"got", lastString(test.input),
			)
		}
	}
}

func TestSwapFileExtension(t *testing.T) {
	type swapTests struct {
		input        string
		newExtension string
		expected     string
	}

	var tests = []swapTests{
		{"some/where/foo_bar.baz", "csv", "some/where/foo_bar.csv"},
		{"foo_bar.baz", "json", "foo_bar.json"},
	}

	for _, test := range tests {
		result := swapFileExtension(test.input, test.newExtension)
		if result != test.expected {
			t.Error(
				"Got", result, "Expected", test.expected,
			)
		}
	}
}
