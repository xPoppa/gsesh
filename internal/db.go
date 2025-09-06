package db

import (
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
	defer db.Close()
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
	db.db.Close()
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
		bpid := b.Get([]byte(key))
		if bpid != nil {
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
	return Result{Pid: pid, Exists: exists}, nil
}
