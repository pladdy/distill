// BoltDB struct to act as a layer between boltDB and any code using this k/v
// store.  Idea of having this is if boltDB is no longer the deisred k/v store
// this layer can be swapped out as long as the API is preserved.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pladdy/lumberjack"
)

type BoltDB struct {
	Db     *bolt.DB
	DbPath string
	DbName string
	Bucket string
}

// Similar to Give, but Append will do a take on a key first and if not nil
// append it's value to the key instead of replacing
func (store *BoltDB) Append(key string, value string) {
	store.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(store.Bucket))
		storedValue := bucket.Get([]byte(key))
		var err error

		if storedValue == nil {
			err = bucket.Put([]byte(key), []byte(value))
		} else {
			newValue := string(storedValue) + "\n" + value
			err = bucket.Put([]byte(key), []byte(newValue))
		}

		return err
	})
}

// BoltDb batch is only useful with goroutines calling it.  This takes they
// key and value to append along with a channel.  It sends it's result to the
// channel to get handled elsewhere
func (store *BoltDB) BatchAppend(key string, value string, c chan error) {
	c <- store.Db.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(store.Bucket))
		storedValue := bucket.Get([]byte(key))
		var err error

		if storedValue == nil {
			err = bucket.Put([]byte(key), []byte(value))
		} else {
			newValue := string(storedValue) + "\n" + value
			err = bucket.Put([]byte(key), []byte(newValue))
		}

		return err
	})
}

// Start the KV Store and by creating the db
func (store *BoltDB) Create(dbPath string) {
	lumberjack.StartLogging()
	lumberjack.Info("Creating db named " + dbPath)

	store.DbPath = dbPath

	var err error
	store.Db, err = bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		lumberjack.Fatal("Couldn't create a db at %v: %v", dbPath, err)
	}
}

func (store *BoltDB) Close() {
	lumberjack.Info("Closing %v", store.DbPath)
	store.Db.Close()
}

// Close and Remove the db from the file system
func (store *BoltDB) Destroy() {
	store.Close()
	lumberjack.Warn("Removing db file " + store.DbPath)
	os.Remove(store.DbPath)
}

// Give data to a key in the BoltDB's bucket; overwrites value if there
func (store *BoltDB) Give(key string, value string) {
	store.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(store.Bucket))
		err := bucket.Put([]byte(key), []byte(value))
		return err
	})
}

// Given a name, use it as the current Bucket
func (store *BoltDB) SetBucket(bucketName string) {
	store.Bucket = bucketName
	lumberjack.Info("Modifying bucket " + store.Bucket)

	store.Db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(store.Bucket))
		if err != nil {
			lumberjack.Fatal("Failed to create a bucket " + store.Bucket)
		}

		rightNow := fmt.Sprintf("%v", time.Now())
		lumberjack.Info("Setting 'Bucket Modification Time' to " + rightNow)

		err = bucket.Put([]byte("Bucket Modification Time"), []byte(rightNow))
		return err
	})
}

// Take data from a key in the BoltDB's bucket
func (store *BoltDB) Take(key string) (value []byte) {
	store.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(store.Bucket))
		value = bucket.Get([]byte(key))
		return nil
	})

	return value
}

// Take data from a key in the BoltDB's bucket as a string
func (store *BoltDB) TakeString(key string) (value string) {
	return string(store.Take(key))
}
