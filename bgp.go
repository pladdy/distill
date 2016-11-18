// BGP Converter to convert bgpdumps to json files
// http://bgp.us/ for more info on BGP
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pladdy/lumberjack"
)

// Example BGP Record and indexes below
// TABLE_DUMP2|1474983369|B|212.25.27.44|8758|0.0.0.0/0|8758 6830|IGP|212.25.27.44|0|0 |8758:110 8758:300|NAG| |
// 0           1          2 3            4    5          6        7    8           9 10 11                12  13

const (
	defaultBucketName = "BGP"
	// tried increasing this to get more performance but it just slows down;
	// also tried increasing this and increasing the max batches in boltdb but
	// still slowed down overall
	maxBatches            = 1000
	updateAfterProcessing = 100000
	joinString            = "|"
	// record indexes
	timeIndex    = 1
	fromIPIndex  = 3
	fromASNIndex = 4
	prefixIndex  = 5
	pathIndex    = 6
)

type seenASPath struct {
	ModificationTime     int
	FromIP               string
	FromASN              int
	Prefix               string
	AutonomousSystemPath []int
}

type bgpRecord struct {
	AutonomousSystem      int
	AutonomousSystemPaths []seenASPath
	Prefixes              []string
}

// Given a bgpdump file, convert it to csv
func bgpDumpToCSV(sourceFile string) string {
	csvDumpFile := filepath.Base(swapFileExtension(sourceFile, "csv"))
	lumberjack.Info("Running `bgpdump` on file %v to %v", sourceFile, csvDumpFile)

	dumpCmd := exec.Command(
		"bgpdump",
		"-t",
		"change",
		"-m",
		"-O",
		csvDumpFile,
		sourceFile)

	var out bytes.Buffer
	dumpCmd.Stdout = &out
	err := dumpCmd.Run()
	if err != nil {
		lumberjack.Fatal(
			"Failed to bgpdump file %v: Err: %v, Out: %q",
			sourceFile,
			err,
			out.String())
	}
	return csvDumpFile
}

// Given a file created with `bgpdump -t`, convert it to aa JSON format grouped
// by the AS in the record being observed
func distillBGP(sourceFile string, destinationFile string) {
	var store BoltDB
	store.Create(swapFileExtension(destinationFile, "db"))
	store.SetBucket(defaultBucketName)
	defer store.Destroy()

	csvDumpFile := bgpDumpToCSV(sourceFile)
	fileHandle, err := os.Open(csvDumpFile)
	if err != nil {
		lumberjack.Fatal("Failed to open file to read")
	}
	defer fileHandle.Close()

	csvReader := csv.NewReader(fileHandle)
	csvReader.Comma = '|'

	start := time.Now()
	records, asMap := storeBGP(csvReader, &store)
	lumberjack.Info("Processed %.f total records in %v", records, time.Since(start))

	dumpBGPStore(&store, asMap, destinationFile)
}

// Dump the stored bgp data to a json file
func dumpBGPStore(store *BoltDB, asMap map[string]int, fileName string) {
	lumberjack.Info("Dumping store to file %v", fileName)

	file, err := os.Create(fileName)
	if err != nil {
		lumberjack.Error("Failed to create file %v", file)
	}
	defer file.Close()

	lastIndex := len(asMap) - 1
	index := 0
	file.Write([]byte("[\n"))

	for asn := range asMap {
		records := store.TakeString(asn)

		data, err := marshalBGP(asn, records)
		if err != nil {
			lumberjack.Error("Failed to marshal records %v:\n%v", records, err)
		}
		file.Write([]byte("  "))
		file.Write(data)
		if index != lastIndex {
			file.Write([]byte(",\n"))
		}
		index += 1
	}
	file.Write([]byte("\n]"))
}

// Given a record from a bgpdump, expand the AS path if there's an AS set
// and return the record
func expandASPath(asPath string) (expandedPath string) {
	itMatched, err := regexp.MatchString("{", asPath)
	if err != nil {
		lumberjack.Error("regexp.Match failed: %v", err)
	}

	// if we have a match then there's an as set to expand; expand and return it
	if itMatched == true {
		expandedPath = strings.Replace(asPath, "{", "", -1)
		expandedPath = strings.Replace(expandedPath, "}", "", -1)
		expandedPath = strings.Replace(expandedPath, ",", " ", -1)
		return
	}
	return asPath
}

// Given a string of newline separated bgp records, marshal them into json
func marshalBGP(asn string, records string) ([]byte, error) {
	asNumber, err := strconv.Atoi(asn)
	if err != nil {
		return []byte(nil), err
	}

	bgp := bgpRecord{
		AutonomousSystem:      asNumber,
		AutonomousSystemPaths: systemPaths(records),
		Prefixes:              uniquePrefixes(records),
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
			lumberjack.Error("Error in csv read; Err: %s, Record: %s", err, record)
			continue
		}

		// track stats and prepare record
		record[pathIndex] = expandASPath(record[pathIndex])
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
			provideUpdate(&stopWatch, updateAfterProcessing, recordsSoFar)
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
			asn, err := strconv.Atoi(as)
			if err != nil {
				lumberjack.Warn("Failed to convert AS %v; %v", as, err)
			} else {
				asns = append(asns, asn)
			}
		}

		modTime, err := strconv.Atoi(values[timeIndex])
		if err != nil {
			lumberjack.Warn("Failed to convert %v; %v", values[timeIndex], err)
		}
		fromASN, err := strconv.Atoi(values[fromASNIndex])
		if err != nil {
			lumberjack.Warn("Failed to convert %v; %v", values[fromASNIndex], err)
		}

		paths = append(paths, seenASPath{
			ModificationTime:     modTime,
			FromIP:               string(values[fromIPIndex]),
			FromASN:              fromASN,
			Prefix:               string(values[prefixIndex]),
			AutonomousSystemPath: asns,
		})
	}

	return paths
}

// Given a string of BGP records, return a unique list of prefixes
func uniquePrefixes(records string) (uniques []string) {
	seenPrefixes := make(map[string]bool)

	for _, record := range strings.Split(records, "\n") {
		values := strings.Split(record, joinString)
		prefix := values[prefixIndex]

		if seenPrefixes[prefix] == false {
			uniques = append(uniques, prefix)
			seenPrefixes[prefix] = true
		}
	}

	return
}
