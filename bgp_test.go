package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pladdy/lumberjack"
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

func TestSystemPaths(t *testing.T) {
	lumberjack.StartLogging()

	// create a record with '{' and '}' in them
	oddRecord := make([]string, len(rawRecord))
	copy(oddRecord, rawRecord)
	oddRecord[6] = "1234 5678 {357, 2124}"

	testRecords := rebuildRecord(rawRecord)
	testRecords = testRecords + "\n" + rebuildRecord(oddRecord)

	fmt.Println(systemPaths(testRecords))
}
