package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/pladdy/lumberjack"
)

const (
	defaultBucketName = "BGP"
	maxBatches        = 1000
)

var bgpIndexesToKeep = []int{1, 5, 6}

// Given a file created with `bgpdump -t`, convert it to aa JSON format grouped
// by the AS in the record being observed
func distillBGP(file string) {
	var store BoltDB
	store.Create(dbNameFromFile(file))
	store.SetBucket(defaultBucketName)
	defer store.Destroy()

	fileHandle, err := os.Open(file)
	if err != nil {
		lumberjack.Fatal("Failed to open file to read")
	}
	defer fileHandle.Close()

	csvReader := csv.NewReader(fileHandle)
	csvReader.Comma = '|'

	processStart := time.Now()
	recordsStored, asMap := storeBGP(csvReader, &store)
	lumberjack.Info("Processed %.f total records in %v", recordsStored, time.Since(processStart))
	fmt.Println(asMap["1"])
}

// For each record in the file, store the record according to the last AS in
// the AS path; this is the AS being reported on
func storeBGP(csvReader *csv.Reader, store *BoltDB) (float64, map[string]int) {
	recordsSoFar := 0.0
	asns := make(map[string]int)
	stopWatch := time.Now()
	batchChannel := make(chan error, maxBatches)

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			lumberjack.Warn("Error in csv read; Err: %s, Record: %s", err, record)
			continue
		}

		// track stats and prepare record
		record = filterBGPRecord(record, bgpIndexesToKeep)
		asn := lastString(strings.Split(record[2], " "))
		if asn == "" {
			lumberjack.Warn("No key for record %v", record)
			continue
		}
		asns[asn]++
		recordsSoFar++

		// batch write the record; drain it if we hit the maxBatches
		go store.BatchAppend(asn, strings.Join(record, "|"), batchChannel)
		if math.Mod(recordsSoFar, maxBatches) == 0 {
			drainChannel(batchChannel, maxBatches)
		}

		// log processing update if necessary
		if math.Mod(recordsSoFar, updateAfterProcessing) == 0 {
			provideUpdate(&stopWatch, recordsSoFar)
		}
	}
	return recordsSoFar, asns
}

// Given a BGP csv record, remove the cruft
// TABLE_DUMP2|1474983369|B|212.25.27.44|8758|0.0.0.0/0|8758 6830|IGP|212.25.27.44|0|0 |8758:110 8758:300|NAG| |
// 0           1          2 3            4    5          6        7    8           9 10 11                12  13
func filterBGPRecord(bgpRecord []string, indexes []int) []string {
	var keepers []string
	for _, index := range indexes {
		keepers = append(keepers, bgpRecord[index])
	}
	return keepers
}
