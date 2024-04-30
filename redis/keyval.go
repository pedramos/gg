package main

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	KeyValEngineType = reflect.TypeFor[KeyVal]()
)

type StorageEngine interface {
	Set(key string, value any) error
	Get(key string) (any, error)
}

func NewStorageEngine(t reflect.Type) (StorageEngine, error) {
	switch t {
	case reflect.TypeFor[KeyVal]():
		return NewKeyVal(), nil
	default:
		return nil, fmt.Errorf("unsupported storage engine")
	}
}

type KeyVal struct {
	sync.RWMutex
	storage map[string]any
}

func NewKeyVal() *KeyVal {
	return &KeyVal{
		storage: make(map[string]any),
	}
}

func (kv *KeyVal) Set(key string, value any) error {
	kv.Lock()
	defer kv.Unlock()

	kv.storage[key] = value
	return nil
}

func (kv *KeyVal) Get(key string) (any, error) {
	kv.RLock()
	defer kv.RUnlock()
	v := kv.storage[key]
	return v, nil
}
