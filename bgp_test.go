package main

import (
	"strings"
	"testing"
)

type recordTests struct {
	indexes        []int
	rawRecord      []string
	expectedRecord []string
}

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

func rebuildRecord(pieces []string) string {
	return strings.Join(pieces, "|")
}

// Given a BGP csv record, remove the cruft
// TABLE_DUMP2|1474983369|B|212.25.27.44|8758|0.0.0.0/0|8758 6830|IGP|212.25.27.44|0|0 |8758:110 8758:300|NAG| |
// 0           1          2 3            4    5          6        7    8           9 10 11                12  13
func TestFilterBGPRecord(t *testing.T) {
	var recordTests = []recordTests{
		{[]int{0}, rawRecord, []string{rawRecord[0]}},
		{[]int{0, 1}, rawRecord, []string{rawRecord[0], rawRecord[1]}},
		{[]int{1, 3, 4, 5, 6}, rawRecord, []string{
			rawRecord[1],
			rawRecord[3],
			rawRecord[4],
			rawRecord[5],
			rawRecord[6],
		}},
		{[]int{1, 5, 6}, rawRecord, []string{
			rawRecord[1],
			rawRecord[5],
			rawRecord[6],
		}},
	}

	for _, triple := range recordTests {
		filteredRecord := filterBGPRecord(triple.rawRecord, triple.indexes)

		if rebuildRecord(filteredRecord) != rebuildRecord(triple.expectedRecord) {
			t.Error(
				"For", rawRecord,
				"expected", triple.expectedRecord,
				"got", filteredRecord,
			)
		}
	}
}
