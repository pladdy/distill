package main

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pladdy/lumberjack"
)

// Example BGP Record and indexes below
// TABLE_DUMP2|1474983369|B|212.25.27.44|8758|0.0.0.0/0|8758 6830|IGP|212.25.27.44|0|0 |8758:110 8758:300|NAG| |
// 0           1          2 3            4    5          6        7    8           9 10 11                12  13

const (
	defaultBucketName     = "BGP"
	maxBatches            = 1000
	updateAfterProcessing = 100000
	prefixIndex           = 5
	pathIndex             = 6
)

type seenASPath struct {
	ModificationTime     int
	AutonomousSystemPath []int
}

type bgpRecord struct {
	AutonomousSystem      int
	AutonomousSystemPaths []seenASPath
	Prefixes              []string
}

// Given a file created with `bgpdump -t`, convert it to aa JSON format grouped
// by the AS in the record being observed
func distillBGP(file string) {
	var store BoltDB
	store.Create(swapFileExtension(file, "db"))
	store.SetBucket(defaultBucketName)
	defer store.Destroy()

	fileHandle, err := os.Open(file)
	if err != nil {
		lumberjack.Fatal("Failed to open file to read")
	}
	defer fileHandle.Close()

	csvReader := csv.NewReader(fileHandle)
	csvReader.Comma = '|'

	start := time.Now()
	records, asMap := storeBGP(csvReader, &store)
	lumberjack.Info("Processed %.f total records in %v", records, time.Since(start))

	dumpBGP(&store, asMap, swapFileExtension(file, "json"))
}

// Dump the stored bgp data to a json file
func dumpBGP(store *BoltDB, asMap map[string]int, fileName string) {
	lumberjack.Info("Dumping store to file %v", fileName)

	file, err := os.Create(fileName)
	if err != nil {
		lumberjack.Error("Failed to create file %v", file)
	}
	defer file.Close()

	file.Write([]byte("["))
	file.Write([]byte("\n  "))

	for asn := range asMap {
		records := store.TakeString(asn)

		data, err := marshalBGP(asn, records)
		if err != nil {
			lumberjack.Error("Failed to marshal records %v:\n%v", records, err)
		}
		file.Write(data)
		file.Write([]byte("\n"))
	}
	file.Write([]byte("]"))
}

// Given a string of newline separated bgp records, marshal them into json
func marshalBGP(asn string, records string) ([]byte, error) {
	asNumber, err := strconv.Atoi(asn)
	if err != nil {
		return []byte(nil), err
	}

	var prefixes []string
	for _, record := range strings.Split(records, "\n") {
		values := strings.Split(record, "|")
		prefixes = append(prefixes, values[prefixIndex])
	}

	bgp := bgpRecord{
		AutonomousSystem:      asNumber,
		AutonomousSystemPaths: systemPaths(records),
		Prefixes:              uniquePrefixes(prefixes),
	}

	return json.MarshalIndent(bgp, "  ", "  ")
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
		asn := lastString(strings.Split(record[pathIndex], " "))
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

// Given a string of BGP records, return a list of seenASPath structs
func systemPaths(records string) []seenASPath {
	var paths []seenASPath

	for _, record := range strings.Split(records, "\n") {
		values := strings.Split(record, "|")

		var asns []int
		for _, as := range strings.Split(values[pathIndex], " ") {
			as = strings.Replace(as, "{", "", -1)
			as = strings.Replace(as, "}", "", -1)

			asn, err := strconv.Atoi(as)
			if err != nil {
				lumberjack.Warn("Failed to convert AS %v; %v", as, err)
			} else {
				asns = append(asns, asn)
			}
		}

		modTime, err := strconv.Atoi(values[0])
		if err == nil {
			paths = append(paths, seenASPath{
				ModificationTime:     modTime,
				AutonomousSystemPath: asns,
			})
		}
	}

	return paths
}

// Given a string of BGP records, return a unique list of prefixes
func uniquePrefixes(prefixes []string) []string {
	seenPrefixes := make(map[string]bool)
	var uniques []string

	for _, prefix := range prefixes {
		if seenPrefixes[prefix] == false {
			uniques = append(uniques, prefix)
			seenPrefixes[prefix] = true
		}
	}

	return uniques
}
