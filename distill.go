package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/pladdy/lumberjack"
)

func main() {
	lumberjack.StartLogging()

	if len(os.Args) == 1 {
		usage()
    lumberjack.Fatal("Missing arguments: need a source file and destination file")
	}

	distillBGP(os.Args[1], os.Args[2])
}

// Given a channel and a number of times to take from it, attempt to take from
// it that many times
func drainChannel(c chan error, times int) {
	for i := 0; i < times; i++ {
		if err := <-c; err != nil {
			lumberjack.Fatal("Failed to batch! %v", err)
		}
	}
}

// Given a slice of strings, return the last one
func lastString(strings []string) string {
	return strings[len(strings)-1]
}

// Given a pointer to a time.Time, and a count of records, provide a logged
// update; tells you how long it took to process the records so far
func provideUpdate(stopWatch *time.Time, newRecords float64, totalRecords float64) {
	lumberjack.Info(
		"Processed %.f records of %.f in %v",
		newRecords,
		totalRecords,
		time.Since(*stopWatch))
	*stopWatch = time.Now()
}

// Given a file name and new extension, swap the extensions
func swapFileExtension(fileName string, extension string) string {
	re := regexp.MustCompile("\\" + filepath.Ext(fileName) + "$")
	return string(
		re.ReplaceAll([]byte(fileName), []byte("."+extension)))
}

func usage() {
	fmt.Println("call with <file to distil> <file to distill to>")
}
