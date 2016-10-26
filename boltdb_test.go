package main

import (
	"os"
	"testing"

	"github.com/pladdy/lumberjack"
)

var (
	testBoltDB BoltDB
	testDBPath string = "./testBolt.db"
)

func setup() {
	lumberjack.Hush()
	testBoltDB.Create(testDBPath)
}

func teardown() {
	testBoltDB.Destroy()
}

func TestCreate(t *testing.T) {
	setup()
	defer teardown()
	_, err := os.Stat(testDBPath)

	if err != nil || os.IsNotExist(err) {
		t.Error("Failed to create db file")
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
