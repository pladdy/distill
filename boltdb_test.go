package main

import (
	"os"
	"testing"

	"github.com/pladdy/lumberjack"
)

// set up oriented stuff
var (
	testBoltDB BoltDB
	testDBPath string = "./testBolt.db"
)

func setup() {
	lumberjack.Hush()
	testBoltDB.Create(testDBPath)
	testBoltDB.SetBucket("test")
}

func teardown() {
	testBoltDB.Destroy()
}

// tests
func TestAppend(t *testing.T) {
	setup()
	defer teardown()

	testKey := "test"
	testValue := "niner"

	testBoltDB.Append(testKey, testValue)
	result := testBoltDB.TakeString(testKey)
	if result != testValue {
		t.Error("Got:", result, "Expected:", testValue)
	}

	// append something to something that exists
	concatValue := testValue + "\n" + testValue

	testBoltDB.Append(testKey, testValue)
	result = testBoltDB.TakeString(testKey)
	if result != concatValue {
		t.Error("Got:", result, "Expected:", concatValue)
	}
}

func TestBatchAppend(t *testing.T) {
	setup()
	defer teardown()

	testKey := "test"
	testValue := "niner"

	maxAppends := 10
	errorChannel := make(chan error, 10)

	for i := 0; i < maxAppends; i++ {
		go testBoltDB.BatchAppend(testKey, testValue, errorChannel)
	}

	// drain channel and build an expected result
	appendedString := testValue

	for i := 0; i < maxAppends; i++ {
		<-errorChannel // to the ether!

		// don't append the first iteration, we already did that
		if i > 0 {
			appendedString += "\n" + testValue
		}
	}

	result := testBoltDB.TakeString(testKey)
	if result != appendedString {
		t.Error("Got:", result, "Expected:", appendedString)
	}
}

func TestCreate(t *testing.T) {
	setup()
	defer teardown()

	_, err := os.Stat(testDBPath)
	if err != nil || os.IsNotExist(err) {
		t.Error("Failed to create db file")
	}
}

func TestClose(t *testing.T) {
	setup()
	defer teardown()

	testBoltDB.Close()
	testBoltDB.Give("test", "niner")
	result := testBoltDB.TakeString("test")

	if result != "" {
		t.Error("Expected DB to be closed and not allow writes")
	}
}

func TestDestroy(t *testing.T) {
	setup()
	testBoltDB.Destroy()

	_, err := os.Stat(testDBPath)
	if os.IsNotExist(err) == false {
		t.Error("Failed to delete db file")
	}
}

func TestGive(t *testing.T) {
	setup()
	defer teardown()

	testKey := "test"
	testValue := "niner"

	testBoltDB.Give(testKey, testValue)
	result := testBoltDB.TakeString(testKey)
	if result != testValue {
		t.Error("Got:", result, "Expected:", testValue)
	}
}

func TestTake(t *testing.T) {
	setup()
	defer teardown()

	testKey := "test"
	testValue := "niner"

	testBoltDB.Give(testKey, testValue)
	result := testBoltDB.Take(testKey)
	if string(result) != testValue {
		t.Error("Got:", string(result), "Expected:", testValue)
	}
}

func TestTakeString(t *testing.T) {
	setup()
	defer teardown()

	testKey := "test"
	testValue := "niner"

	testBoltDB.Give(testKey, testValue)
	result := testBoltDB.TakeString(testKey)
	if result != testValue {
		t.Error("Got:", result, "Expected:", testValue)
	}
}
