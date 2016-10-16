package main

import (
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pladdy/lumberjack"
)

type KeyValueStore struct {
	Db        *bolt.DB
	DbPath    string
	Container string
}

// Start the KV Store and by creating the db
func (store *KeyValueStore) Create(dbPath string) {
	lumberjack.StartLogging()
	lumberjack.Info("Creating db named " + dbPath)
	store.DbPath = dbPath

	var err error
	store.Db, err = bolt.Open(store.DbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		lumberjack.Fatal("Couldn't create a db file.")
	}
}

// Remove the db from the file system
func (store *KeyValueStore) Destroy() {
	lumberjack.Warn("Removing db file " + store.DbPath)
	store.Db.Close()
	os.Remove(store.DbPath)
}

// Give data to a key in the KeyValueStore's bucket
func (store *KeyValueStore) Give(key string, value string) {
	store.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(store.Container))
		err := bucket.Put([]byte(key), []byte(value))
		return err
	})
}

// Given a name, use it as the current container
func (store *KeyValueStore) SetContainer(bucketName string) {
	store.Container = bucketName
	lumberjack.Info("Modifying bucket " + store.Container)

	store.Db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(store.Container))
		if err != nil {
			lumberjack.Fatal("Failed to create a bucket " + store.Container)
		}

		rightNow := fmt.Sprintf("%v", time.Now())
		lumberjack.Info("Setting 'Bucket Modification Time' to " + rightNow)

		err = bucket.Put([]byte("Bucket Modification Time"), []byte(rightNow))
		return err
	})
}

// Take data from a key in the KeyValueStore's bucket
func (store *KeyValueStore) Take(key string) (value []byte) {
	store.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(store.Container))
		value = bucket.Get([]byte(key))
		return nil
	})

	return value
}

// Take data from a key in the KeyValueStore's bucket as a string
func (store *KeyValueStore) TakeString(key string) (value string) {
	return string(store.Take(key))
}
