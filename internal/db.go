package db

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

const (
	BUCKET_NAME = "sessions"
)

type DB struct {
	db *bolt.DB
}

func NewDB(path string) (*DB, error) {
	db, err := bolt.Open(path, 0660, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BUCKET_NAME))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil

}

func (db *DB) Close() {
	fmt.Println("closing db failed with: ", db.db.Close())
}

func (db *DB) Insert(key string, pid int) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_NAME))
		return b.Put([]byte(key), []byte(strconv.Itoa(pid)))
	})
}

func (db *DB) Delete(key string) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_NAME))
		return b.Delete([]byte(key))
	})
}

type Result struct {
	Pid    int
	Exists bool
}

func (db *DB) GetPid(key string) (Result, error) {
	var pid int
	var exists bool
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_NAME))
		if b == nil {
			return errors.New("Bucket doesn't exist")
		}

		err := b.ForEach(func(k, v []byte) error {
			log.Printf("The key: %s\n", string(k))
			log.Printf("The value: %s\n", string(v))
			return nil
		})
		if err != nil {
			log.Println("logging failed")
		}

		log.Printf("Getting bpid by key: %q", key)
		bpid := b.Get([]byte(key))
		if bpid != nil {
			log.Printf("Bpid exists with: %q", string(bpid))
			exists = true
			p, err := strconv.Atoi(string(bpid))
			if err != nil {
				return err
			}
			pid = p
			return nil
		}

		return nil
	})

	if err != nil {
		return Result{}, err
	}
	log.Printf("Returning the result, exists=%+v, pid:%d", exists, pid)
	return Result{Pid: pid, Exists: exists}, nil
}

type Process struct {
	Key string
	Pid int
}

func (db *DB) ReturnPids() ([]Process, error) {
	var procs []Process
	fmt.Println("db Path", db.db.Path())

	err := db.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(BUCKET_NAME))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			pid, err := strconv.Atoi(string(v))
			if err != nil {
				return err
			}
			procs = append(procs, Process{Key: string(k), Pid: pid})
		}

		return nil
	})

	if err != nil {
		return procs, err
	}

	return procs, nil
}
