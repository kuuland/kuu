package kuu

import (
	"github.com/boltdb/bolt"
)

var cacheBoltDB *bolt.DB

func init() {
	if _, exists := C().Get("redis"); exists {
		// 初始化redis
	} else {
		// 初始化bolt
		db, err := bolt.Open("cache.db", 0600, nil)
		if err != nil {
			FATAL(err)
		} else {
			cacheBoltDB = db
		}
	}
}

func releaseCacheDB() {
	if cacheBoltDB != nil {
		_ = cacheBoltDB.Close()
	}
}
