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

func TestExpandASPath(t *testing.T) {
	lumberjack.Hush()

	var tests = []struct {
		input    string
		expected string
	}{
		{input: "1234 5678 {357,2124}", expected: "1234 5678 357 2124"},
		{input: "1234 5678 {357}", expected: "1234 5678 357"},
	}

	for _, test := range tests {
		result := expandASPath(test.input)
		if result != test.expected {
			t.Error("Expected:", test.expected, "Got:", result)
		}
	}
}

func TestMarshallBGP(t *testing.T) {
	asn := lastString(strings.Split(rawRecord[pathIndex], " "))

	// generate result
	result, err := marshalBGP(asn, strings.Join(rawRecord, "|"))
	if err != nil {
		t.Error("Failed to create a JSON record")
	}
	resultToString := string(result)

	expected := `{
    "AutonomousSystem": 6830,
    "AutonomousSystemPaths": [
      {
        "ModificationTime": 1474983369,
        "FromIP": "212.25.27.44",
        "FromASN": 8758,
        "Prefix": "0.0.0.0/0",
        "AutonomousSystemPath": [
          8758,
          6830
        ]
      }
    ],
    "Prefixes": [
      "0.0.0.0/0"
    ]
  }`

	if resultToString != expected {
		t.Error("Got:", resultToString, "Expected:", expected)
	}
}

func TestSystemPaths(t *testing.T) {
	var tests = []struct {
		records  string
		expected []seenASPath
	}{
		{rebuildRecord(rawRecord),
			[]seenASPath{seenASPath{1474983369, "212.25.27.44", 8758, "0.0.0.0/0", []int{8758, 6830}}}},
	}

	for _, test := range tests {
		results := systemPaths(test.records)
		for i, result := range results {
			expected := test.expected[i]
			if result.ModificationTime != expected.ModificationTime {
				t.Error("Got:", result.ModificationTime, "Expected:", expected.ModificationTime)
			}
			if result.FromIP != expected.FromIP {
				t.Error("Got:", result.FromIP, "Expected:", expected.FromIP)
			}
			if result.FromASN != expected.FromASN {
				t.Error("Got:", result.FromASN, "Expected:", expected.FromASN)
			}
			if result.Prefix != expected.Prefix {
				t.Error("Got:", result.Prefix, "Expected:", expected.Prefix)
			}
			for i, asn := range result.AutonomousSystemPath {
				if asn != expected.AutonomousSystemPath[i] {
					t.Error("Got:", asn, "Expected:", expected.AutonomousSystemPath[i])
				}
			}
		}
	}
}

func TestUniquePrefixes(t *testing.T) {
	testRecords := strings.Join(rawRecord, "|")
	testRecords += "\n" + strings.Join(rawRecord, "|")

	uniqueCIDRs := uniquePrefixes(testRecords)
	if len(uniqueCIDRs) > 1 {
		t.Error("Expected length to be 1")
	}

	if uniqueCIDRs[0] != rawRecord[prefixIndex] {
		t.Error("Expected:", rawRecord[prefixIndex], "Got:", uniqueCIDRs[0])
	}
}
