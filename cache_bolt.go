package kuu

import (
	"bytes"
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

func (c *CacheBolt) seek(seek []byte, f func(k, v []byte) bool) (values map[string]string) {
	values = make(map[string]string)
	ERROR(c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(c.generalBucketName)
		if bucket != nil {
			c := bucket.Cursor()
			for k, v := c.Seek(seek); k != nil && f(k, v); k, v = c.Next() {
				values[string(k)] = string(v)
			}
		}
		return nil
	}))
	return
}

// HasPrefix
func (c *CacheBolt) HasPrefix(prefix string) (values map[string]string) {
	if len(prefix) == 0 {
		return
	}
	seek := []byte(prefix)
	return c.seek(seek, func(k, v []byte) bool {
		return bytes.HasPrefix(k, seek)
	})
}

// HasSuffix
func (c *CacheBolt) HasSuffix(suffix string) (values map[string]string) {
	if len(suffix) == 0 {
		return
	}
	seek := []byte(suffix)
	return c.seek(seek, func(k, v []byte) bool {
		return bytes.HasSuffix(k, seek)
	})
}

// Contains
func (c *CacheBolt) Contains(pattern string) (values map[string]string) {
	if len(pattern) == 0 {
		return
	}
	seek := []byte(pattern)
	return c.seek(seek, func(k, v []byte) bool {
		return bytes.Contains(k, seek)
	})
}

// Search
func (c *CacheBolt) Search(basePattern string, filter func(string, string) bool) (values map[string]string) {
	seek := []byte(basePattern)
	return c.seek(seek, func(k, v []byte) bool {
		return filter(string(k), string(v))
	})
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
		bucket := tx.Bucket(c.generalBucketName)
		if bucket != nil {
			for _, key := range keys {
				if err := bucket.Delete([]byte(key)); err != nil {
					return err
				}
			}
		}
		return nil
	}))
	return
}

// DelLike
func (c *CacheBolt) DelLike(keys ...string) {
	if len(keys) == 0 {
		return
	}

	del := func(bucket *bolt.Bucket, k, v []byte) {
		for _, key := range keys {
			sk := []byte(key)
			if bytes.Contains(k, sk) {
				_ = bucket.Delete([]byte(key))
			}
		}
	}

	ERROR(c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(c.generalBucketName)
		if bucket != nil {
			c := bucket.Cursor()

			k, v := c.Seek([]byte(keys[0]))
			for k != nil {
				del(bucket, k, v)
				k, v = c.Next()
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
