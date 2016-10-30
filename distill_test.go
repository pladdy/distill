package main

import "testing"

type stringTests struct {
	input    []string
	expected string
}

func TestLastString(t *testing.T) {
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
		{"sfoo_bar.baz", "json", "foo_bar.json"},
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
