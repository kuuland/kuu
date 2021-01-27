package kuu

import (
	"github.com/jinzhu/gorm"
	"sync"
)

type HookHandler func(args ...interface{}) error

var (
	gormHandlers   = make(map[string][]func(scope *gorm.Scope) error)
	gormHandlersMu sync.RWMutex

	bizHandlers   = make(map[string][]func(scope *Scope) error)
	bizHandlersMu sync.RWMutex
)

// 模型名称:BeforeCreate
func RegisterGormHook(name string, handler func(scope *gorm.Scope) error) {
	gormHandlersMu.Lock()
	defer gormHandlersMu.Unlock()

	if name != "" && handler != nil {
		gormHandlers[name] = append(gormHandlers[name], handler)
	}
}

// 模型名称:BizBeforeCreate
func RegisterBizHook(name string, handler func(scope *Scope) error) {
	bizHandlersMu.Lock()
	defer bizHandlersMu.Unlock()

	if name != "" && handler != nil {
		bizHandlers[name] = append(bizHandlers[name], handler)
	}
}

func execGormHooks(name string, scope *gorm.Scope) error {
	gormHandlersMu.RLock()
	defer gormHandlersMu.RUnlock()

	for _, item := range gormHandlers[name] {
		if err := item(scope); err != nil {
			return err
		}
	}

	return nil
}

func execBizHooks(name string, scope *Scope) error {
	bizHandlersMu.RLock()
	defer bizHandlersMu.RUnlock()

	for _, item := range bizHandlers[name] {
		if err := item(scope); err != nil {
			return err
		}
	}

	return nil
}
