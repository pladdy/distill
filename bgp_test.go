package main

import (
	"strings"
	"testing"

	"github.com/pladdy/lumberjack"
)

var rawRecord []string = []string{
	"TABLE_DUMP2",
	"1474983369",
	"B",
	"212.25.27.44",
	"8758",
	"0.0.0.0/0",
	"8758 6830",
	"IGP",
	"212.25.27.44",
	"0",
	"0",
	"8758:110 8758:300",
	"NAG",
	"",
}

// Helper
func rebuildRecord(pieces []string) string {
	return strings.Join(pieces, joinString)
}

type expandTests struct {
	input    string
	expected string
}

func TestExpandASPath(t *testing.T) {
	lumberjack.Hush()

	tests := []expandTests{
		{input: "1234 5678 {357,2124}", expected: "1234 5678 357 2124"},
		{input: "1234 5678 {357}", expected: "1234 5678 357"},
	}

	for _, test := range tests {
		result := expandASPath(test.input)
		if result != test.expected {
			t.Error("expected", test.expected, "got", result)
		}
	}
}
