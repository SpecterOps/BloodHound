// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"encoding/json"
	"fmt"
	"reflect"

	lru "github.com/hashicorp/golang-lru"
)

// InvalidValueError is an error return type for invalid values
type InvalidValueError struct {
	Type reflect.Type
}

// Error returns an error depending on the type contained in InvalidValueError
func (s *InvalidValueError) Error() string {
	if s.Type == nil {
		return "cache: invalid value passed: nil"
	}

	if s.Type.Kind() != reflect.Pointer {
		return "cache: invalid value passed: non-pointer"
	}

	return fmt.Sprintf("cache: invalid value passed: nil %s", s.Type.String())
}

// Config contains configuration for our cache. This config should contain any
// cache configuration options we actually use.
type Config struct {
	MaxSize int // Max size of cache in number of items
}

// Cache wraps our underlying cache implementation.
type Cache struct {
	lru *lru.Cache
}

func get(cache *lru.Cache, key string, value any) (bool, error) {
	if cachedJSON, ok := cache.Get(key); !ok {
		return false, nil
	} else if err := json.Unmarshal(cachedJSON.([]byte), &value); err != nil {
		return false, fmt.Errorf("error unmarshalling cached entry: %w", err)
	} else {
		return true, nil
	}
}

func set(cache *lru.Cache, key string, value any) (int, bool, error) {
	if cachedJSON, err := json.Marshal(value); err != nil {
		return 0, false, fmt.Errorf("error marshalling value: %w", err)
	} else {
		eviction := cache.Add(key, cachedJSON)
		// Select the size of the cached value to aid in logging
		return len(cachedJSON), eviction, nil
	}
}

// Get takes a key and a pointer to a value and sets the value to the corresponding
// cache entry. Returns true if the key was found, false otherwise. Returns an
// error if the underlying cache returns an error during getting the value or if
// the value couldn't be unmarshalled.
func (s Cache) Get(key string, value any) (bool, error) {
	if rv := reflect.ValueOf(value); rv.Kind() != reflect.Pointer || rv.IsNil() {
		return false, &InvalidValueError{rv.Type()}
	} else {
		return get(s.lru, key, value)
	}
}

// Set takes a key and a value and sets the value in the cache. Returns number
// of bytes written. Returns true if an eviction occured, else false. Returns an
// error if the underlying cache returns an error during setting the value or if
// the value couldn't be marshalled.
func (s Cache) Set(key string, value any) (int, bool, error) {
	return set(s.lru, key, value)
}

// GuardedSet takes a key and a value and sets the value in the cache if it cannot
// be found. Returns true if value was set, false otherwise. Returns number of bytes
// written. Returns an error if the underlying cache returns an error during setting
// the value or if the value couldn't be marshalled.
func (s Cache) GuardedSet(key string, value any) (bool, int, error) {
	if ok, err := get(s.lru, key, value); err != nil {
		return false, 0, fmt.Errorf("error checking cache entry exists: %w", err)
	} else if ok {
		return false, 0, nil
	} else {
		// Currently we don't need to know about evictions with GuardedSet so ignoring
		// to keep interface sane
		bytesWritten, _, err := set(s.lru, key, value)
		return true, bytesWritten, err
	}
}

// Len returns the length of the current cache
func (s Cache) Len() int {
	return s.lru.Len()
}

// Reset attempts to reset the underlying cache. Returns an error if the underlying
// cache returns an error during reset.
func (s Cache) Reset() error {
	s.lru.Purge()
	// This is to provide backwards compatibility for our interface
	return nil
}

// NewCache takes a cache config. Returns a new Cache instance and an error if the underlying
// cache returns an error during configuration.
func NewCache(config Config) (Cache, error) {
	if cache, err := lru.New(config.MaxSize); err != nil {
		return Cache{}, fmt.Errorf("error creating cache: %w", err)
	} else {
		return Cache{lru: cache}, nil
	}
}
