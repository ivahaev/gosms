package sdb

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/ivahaev/bolt-view"
	"github.com/ivahaev/go-logger"
	"sync"
	"errors"
)

var DB *bolt.DB
var opened bool
var mutex = &sync.Mutex{}

type M map[string]interface{}

func Opened() bool {
	mutex.Lock()
	defer mutex.Unlock()
	return opened
}

func init() {
	dbpath := "./sms.db"
	var err error
	DB, err = bolt.Open(dbpath, 0600, nil)
	if err != nil {
		panic("Can't open DB file " + dbpath)
	}
	mutex.Lock()
	defer mutex.Unlock()
	opened = true
	go boltview.Init(DB, "3334")
}

func Delete(bucket, key string) (err error) {
	DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			err = errors.New("Not found")
			return err
		}
		err = b.Delete([]byte(key))
		return err
	})
	return
}

func DeleteBucket(bucket string) (err error) {
	DB.Update(func(tx *bolt.Tx) error {
		err = tx.DeleteBucket([]byte(bucket))
		return err
	})
	return
}

func Get(bucket, key string) (result []byte, err error) {
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			err = errors.New("Not found")
			return err
		}
		v := b.Get([]byte(key))

		if v == nil {
			err = errors.New("Not found")
			return err
		}
		result = append(result, v...)

		return nil
	})
	return
}

func GetAll(bucket string) (result [][]byte, err error) {
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			err = errors.New("Not found")
			return err
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			result = append(result, v)
		}
		return nil
	})
	return
}

func GetAllWithKeys(bucket string) (result map[string][]byte, err error) {
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			err = errors.New("Not found")
			return err
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			result[string(k)] = v
		}
		return nil
	})
	return
}

func GetAllKeys(bucket string) (data []string, err error) {
	data = []string{}
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			err = errors.New("Not found")
			return err
		}
		b.ForEach(func(k, v []byte) error {
			data = append(data, string(k))
			return nil
		})
		return nil
	})
	return
}


func GetStatsForBucket(bucket string) (stat bolt.BucketStats) {
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		stat = b.Stats()
		return nil
	})
	return
}

func NewUUIDv4() string {
	u := [16]byte{}
	_, err := rand.Read(u[:16])
	if err != nil {
		panic(err)
	}

	u[8] = (u[8] | 0x80) & 0xBf
	u[6] = (u[6] | 0x40) & 0x4f

	return fmt.Sprintf("%x-%x-%x-%x-%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func Save(bucket, key string, value interface{}) (err error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			logger.Error("Can't create bucket:", bucket, err.Error())
			return err
		}
		err = b.Put([]byte(key), encoded)
		return err
	})
	return
}

func Set(bucket, key string, value []byte) (err error) {
	DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			logger.Error("Can't create bucket:", bucket, err.Error())
			return err
		}
		err = b.Put([]byte(key), value)
		return err
	})
	return
}
