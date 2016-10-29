package main

import (
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/pladdy/lumberjack"
)

// Given a file name, use the name to return a .db file name
func swapFileExtension(fileName string, extension string) string {
	re := regexp.MustCompile("\\" + filepath.Ext(fileName) + "$")
	return string(
		re.ReplaceAll([]byte(filepath.Base(fileName)), []byte("."+extension)))
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
func provideUpdate(stopWatch *time.Time, records float64) {
	sinceLastUpdate := time.Since(*stopWatch)
	lumberjack.Info("Processed %.f records in %v", records, sinceLastUpdate)
	*stopWatch = time.Now()
}

func main() {
	lumberjack.StartLogging()

	if len(os.Args) == 1 {
		lumberjack.Fatal("File to distill is required as an argument")
	}

	distillBGP(os.Args[1])
}
