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

package cache_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/specterops/bloodhound/cache"
	"github.com/stretchr/testify/require"
)

const (
	testCacheKey1      = "test_cache_key_1"
	testCacheKey2      = "test_cache_key_2"
	unusedTestCacheKey = "test_cache_key_unused"
)

var (
	outputValue testStruct

	invalidInputValue = func() {} // invalid conversion for JSON
	validInputValue1  = testStruct{ID: 0, TestString: "test"}
	validInputValue2  = testStruct{ID: 1, TestString: "test 2"}

	cacheEntries = cacheFillHarness{
		testCacheKey1: validInputValue1,
		testCacheKey2: validInputValue2,
	}
)

type testStruct struct {
	ID         int
	TestString string
}

type cacheFillHarness map[string]testStruct

func getPopulatedInstance(data cacheFillHarness) (cache.Cache, error) {
	if instance, err := cache.NewCache(cache.Config{len(data)}); err != nil {
		return instance, fmt.Errorf("failed to create new cache instance: %w", err)
	} else {
		for key, value := range data {
			if _, evicted, err := instance.Set(key, value); err != nil {
				return instance, fmt.Errorf("failed to set value %+v to key %s: %w", value, key, err)
			} else if evicted {
				return instance, fmt.Errorf("failed to set value %+v to key %s: an unexpected eviction occurred", value, key)
			}
		}

		return instance, nil
	}
}

func TestCache_NewCache(t *testing.T) {
	var (
		testValidSizes = []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096}
	)

	t.Run("NewCache should fail if MaxSize is 0", func(t *testing.T) {
		_, err := cache.NewCache(cache.Config{MaxSize: 0})
		require.NotNil(t, err)
	})

	for _, size := range testValidSizes {
		t.Run(fmt.Sprintf("NewCache should succeed if MaxSize is: %d", size), func(t *testing.T) {
			_, err := cache.NewCache(cache.Config{MaxSize: 1})
			require.Nil(t, err)
		})
	}
}

func TestCache_Set(t *testing.T) {
	_, jsonErr := json.Marshal(invalidInputValue)
	require.NotNil(t, jsonErr)

	jsonValidInputValue1, err := json.Marshal(validInputValue1)
	require.Nil(t, err)

	jsonValidInputValue2, err := json.Marshal(validInputValue2)
	require.Nil(t, err)

	t.Run("Set using invalid value fails", func(t *testing.T) {
		instance, err := cache.NewCache(cache.Config{1})
		require.Nil(t, err)

		_, _, err = instance.Set(testCacheKey1, &invalidInputValue)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), jsonErr.Error())
	})

	t.Run("Set using valid value succeeds", func(t *testing.T) {
		instance, err := cache.NewCache(cache.Config{1})
		require.Nil(t, err)

		written, eviction, err := instance.Set(testCacheKey1, validInputValue1)
		require.Nil(t, err)
		require.False(t, eviction)
		require.Equal(t, len(jsonValidInputValue1), written)
	})

	t.Run("Set using the same key should overwrite existing value", func(t *testing.T) {
		instance, err := getPopulatedInstance(cacheEntries)
		require.Nil(t, err)

		written, eviction, err := instance.Set(testCacheKey1, validInputValue2)
		require.Nil(t, err)
		require.False(t, eviction)
		require.Equal(t, len(jsonValidInputValue2), written)

		ok, err := instance.Get(testCacheKey1, &outputValue)
		require.Nil(t, err)
		require.True(t, ok)
		require.EqualValues(t, validInputValue2, outputValue)
	})
}

func TestCache_GuardedSet(t *testing.T) {
	_, jsonErr := json.Marshal(invalidInputValue)
	require.NotNil(t, jsonErr)

	json, err := json.Marshal(validInputValue1)
	require.Nil(t, err)

	t.Run("GuardedSet using invalid value fails", func(t *testing.T) {
		instance, err := cache.NewCache(cache.Config{1})
		require.Nil(t, err)

		_, _, err = instance.GuardedSet(testCacheKey1, invalidInputValue)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), jsonErr.Error())
	})

	t.Run("GuardedSet using existing key returns false with 0 bytes written", func(t *testing.T) {
		instance, err := cache.NewCache(cache.Config{1})
		require.Nil(t, err)

		_, _, err = instance.Set(testCacheKey1, validInputValue1)
		require.Nil(t, err)

		ok, written, err := instance.GuardedSet(testCacheKey1, validInputValue1)
		require.Nil(t, err)
		require.False(t, ok)
		require.Equal(t, 0, written)
	})

	t.Run("GuardedSet using unique key writes to cache", func(t *testing.T) {
		instance, err := cache.NewCache(cache.Config{1})
		require.Nil(t, err)

		ok, written, err := instance.GuardedSet(testCacheKey2, validInputValue1)
		require.Nil(t, err)
		require.True(t, ok)
		require.Equal(t, len(json), written)
	})
}

func TestCache_Get(t *testing.T) {
	instance, err := getPopulatedInstance(cacheEntries)
	require.Nil(t, err)

	t.Run("Get using non-pointer fails", func(t *testing.T) {
		ok, err := instance.Get(testCacheKey1, outputValue)
		require.NotNil(t, err)
		require.False(t, ok)
		require.Contains(t, err.Error(), "cache: invalid value passed: non-pointer")
	})

	t.Run("Get using pointer succeeds", func(t *testing.T) {
		ok, err := instance.Get(testCacheKey1, &outputValue)
		require.Nil(t, err)
		require.True(t, ok)
		require.EqualValues(t, cacheEntries[testCacheKey1], outputValue)
	})

	t.Run("Get using nonexistent key succeeds with false", func(t *testing.T) {
		ok, err := instance.Get(unusedTestCacheKey, &outputValue)
		require.Nil(t, err)
		require.False(t, ok)
	})
}

func TestCache_Reset(t *testing.T) {
	instance, err := getPopulatedInstance(cacheEntries)
	require.Nil(t, err)

	err = instance.Reset()
	require.Nil(t, err)

	ok, err := instance.Get(testCacheKey1, &outputValue)
	require.Nil(t, err)
	require.False(t, ok)
}

func TestInvalidValueError(t *testing.T) {
	var (
		testValue              = struct{}{}
		invalidValueNil        = cache.InvalidValueError{nil}
		invalidValueNonPointer = cache.InvalidValueError{reflect.ValueOf(testValue).Type()}
		invalidValueCatchAll   = cache.InvalidValueError{reflect.ValueOf(&testValue).Type()}
	)

	require.Equal(t, "cache: invalid value passed: nil", invalidValueNil.Error())
	require.Equal(t, "cache: invalid value passed: non-pointer", invalidValueNonPointer.Error())
	require.Equal(t, "cache: invalid value passed: nil *struct {}", invalidValueCatchAll.Error())
}

func TestCache_FillFifo(t *testing.T) {
	var (
		cacheKeys  = []string{"0", "1", "2", "3", "4", "5"}
		value      testStruct
		testString = "Hello World!"
	)

	instance, err := cache.NewCache(cache.Config{MaxSize: len(cacheKeys)})
	require.Nil(t, err)

	for idx, key := range cacheKeys {
		_, eviction, err := instance.Set(key, testStruct{ID: idx, TestString: testString})
		require.Nil(t, err)
		require.False(t, eviction)
	}

	for idx, key := range cacheKeys {
		// We should be able to retrieve each value from the cache since we never evicted
		ok, err := instance.Get(key, &value)
		require.Nil(t, err)
		require.True(t, ok)
		require.EqualValues(t, testStruct{ID: idx, TestString: testString}, value)
	}

	// Adding a new item once full should cause an eviction and cause the first item to drop out since it was the least recently fetched
	_, eviction, err := instance.Set("overflow_key", testStruct{ID: len(cacheKeys), TestString: testString})
	require.Nil(t, err)
	require.True(t, eviction)

	ok, err := instance.Get("0", &value)
	require.Nil(t, err)
	require.False(t, ok)
}
