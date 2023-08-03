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
	"fmt"
	"log"
	"math"
	"testing"

	"github.com/specterops/bloodhound/cache"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type OUCacheEntry struct {
	ID                graph.ID
	Name              string
	DistinguishedName string
}

const numSimulatedOUs = 500

func getObjectIDs(num int) []string {
	var objectIDs = make([]string, 0, num)
	for i := 0; i < num; i++ {
		objectIDs = append(objectIDs, fmt.Sprintf("object-id-%v", num))
	}
	return objectIDs
}

func setupLRUCache() cache.Cache {
	if c, err := cache.NewCache(cache.Config{MaxSize: numSimulatedOUs}); err != nil {
		log.Fatalf("Error creating cache: %v", err)
	} else {
		return c
	}
	return cache.Cache{}
}

func getEntries(ouCache cache.Cache) {
	var (
		cachedOU OUCacheEntry
	)

	for _, objectID := range getObjectIDs(numSimulatedOUs) {
		ouCache.Get(objectID, &cachedOU)
	}
}

func setEntries(ouCache cache.Cache, entries []string) {
	for id, entry := range entries {
		cacheEntry := OUCacheEntry{
			ID:                graph.ID(id),
			Name:              fmt.Sprintf("Name %v", id),
			DistinguishedName: fmt.Sprintf("Distinguished Name %v", id),
		}
		ouCache.Set(entry, cacheEntry)
	}
}

func BenchmarkOUCacheEmptyLRUCache(b *testing.B) {
	b.ReportAllocs()
	ouCache := setupLRUCache()
	for n := 0; n < b.N; n++ {
		getEntries(ouCache)
	}
}

func BenchmarkOUCacheFillLRUCache(b *testing.B) {
	b.ReportAllocs()
	ouCache := setupLRUCache()
	for n := 0; n < b.N; n++ {
		setEntries(ouCache, getObjectIDs(numSimulatedOUs))
	}
}

func BenchmarkOUCache90PercentLRUCache(b *testing.B) {
	b.ReportAllocs()
	ouCache := setupLRUCache()
	objectIDs := getObjectIDs(numSimulatedOUs)
	numObjects := len(objectIDs)
	first90 := math.Floor(0.9 * float64(numObjects))

	setEntries(ouCache, objectIDs[:int(first90)])
	for n := 0; n < b.N; n++ {
		getEntries(ouCache)
	}
}

func BenchmarkOUCache100PercentLRUCache(b *testing.B) {
	b.ReportAllocs()
	ouCache := setupLRUCache()
	setEntries(ouCache, getObjectIDs(numSimulatedOUs))
	for n := 0; n < b.N; n++ {
		getEntries(ouCache)
	}
}
