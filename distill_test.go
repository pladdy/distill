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
