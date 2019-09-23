package kuu

import (
	"github.com/boltdb/bolt"
	"time"
)

// CacheBolt
type CacheBolt struct {
	db                *bolt.DB
	generalBucketName []byte
}

// NewCacheBolt
func NewCacheBolt() *CacheBolt {
	db, err := bolt.Open("cache.db", 0600, nil)
	if err != nil {
		FATAL(err)
	}
	return &CacheBolt{db, []byte("general")}
}

// SetString
func (c *CacheBolt) SetString(key, val string, expiration ...time.Duration) {
	ERROR(c.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(c.generalBucketName)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(key), []byte(val))
	}))
}

// GetString
func (c *CacheBolt) GetString(key string) (val string) {
	ERROR(c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(c.generalBucketName)
		if bucket != nil {
			val = string(bucket.Get([]byte(key)))
		}
		return nil
	}))
	return
}

// SetInt
func (c *CacheBolt) SetInt(key string, val int, expiration ...time.Duration) {
	ERROR(c.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(c.generalBucketName)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(key), itob(val))
	}))
}

// GetInt
func (c *CacheBolt) GetInt(key string) (val int) {
	ERROR(c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(c.generalBucketName)
		if bucket != nil {
			val = btoi(bucket.Get([]byte(key)))
		}
		return nil
	}))
	return
}

// Incr
func (c *CacheBolt) Incr(key string) (val int) {
	ERROR(c.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(c.generalBucketName)
		if err != nil {
			return err
		}
		v, err := bucket.NextSequence()
		if err != nil {
			return err
		}
		val = int(v)
		return nil
	}))
	return
}

// Del
func (c *CacheBolt) Del(keys ...string) {
	if len(keys) == 0 {
		return
	}
	ERROR(c.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(c.generalBucketName)
		if err != nil {
			return err
		}
		for _, key := range keys {
			if err := bucket.Delete([]byte(key)); err != nil {
				return err
			}
		}
		return nil
	}))
	return
}

// Close
func (c *CacheBolt) Close() {
	if c.db != nil {
		ERROR(c.db.Close())
	}
}
