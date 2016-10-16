package main

import (
	"fmt"

	"github.com/pladdy/lumberjack"
)

func main() {
	lumberjack.StartLogging()

	var store KeyValueStore
	store.Create("ripe00.db")
	defer store.Destroy()

	store.SetContainer("Ripe00")
	store.Give("ASN", "1234")

	lumberjack.Info("Asn Value is " + string(store.Take("ASN")))
	lumberjack.Info("Bucket Modification Time" + store.TakeString("Bucket Modification Time"))

	notThere := string(store.Take("Farts"))
	if notThere == "" {
		fmt.Println("Yep, not there")
	}
}
