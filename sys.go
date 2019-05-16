package kuu

import "sync"

var (
	rootUser  *User
	valuesMap sync.Map
)

// RootUID
func RootUID() uint {
	if rootUser != nil {
		return rootUser.ID
	}
	return 0
}

// RootUser
func RootUser() *User {
	return rootUser
}

// InstantSet
func InstantSet(key, value interface{}) {
	valuesMap.Store(key, value)
}

// InstantGet
func InstantGet(key interface{}) (value interface{}, ok bool) {
	value, ok = valuesMap.Load(key)
	return
}

func getRootUser() *User {
	var root User
	if errs := DB().Model(&User{Username: "root"}).First(&root).GetErrors(); len(errs) > 0 {
		ERROR(errs)
	}
	return &root
}

func createRootUser() *User {
	root := User{
		Username:  "root",
		Name:      "预置用户",
		Password:  MD5("kuu"),
		IsBuiltIn: true,
	}
	if errs := DB().Create(&root).GetErrors(); len(errs) > 0 {
		ERROR(errs)
		PANIC("create root user failed")
	}
	return &root
}

func initSys() {
	root := getRootUser()
	if !DB().NewRecord(root) {
		rootUser = root
	} else {
		rootUser = createRootUser()
	}
}

// Sys
func Sys() *Mod {
	return &Mod{
		Models: []interface{}{
			&User{},
			//&Org{},
		},
		AfterImport: initSys,
	}
}
