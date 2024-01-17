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

package size

import (
	"unsafe"
)

type Sizable interface {
	// SizeOf returns the size of this entity.
	SizeOf() Size
}

func Of[T any](raw T) Size {
	return Size(unsafe.Sizeof(raw))
}

func OfSlice[T comparable](raw []T) Size {
	var template T
	return Of(raw) + Of(template)*Size(cap(raw))
}

func OfContents[T any](raw T) Size {
	switch typed := any(raw).(type) {
	case string:
		return Size(len(typed))
	default:
		return 0
	}
}

type mapHeader struct {
	_type unsafe.Pointer
	value unsafe.Pointer
}

func mapTypeAndValue(mapInstance any) (*maptype, *hmap) {
	// This is filthy
	header := (*mapHeader)(unsafe.Pointer(&mapInstance))
	return (*maptype)(header._type), (*hmap)(header.value)
}

func OfMap[K comparable, V any](mapInstance map[K]V) Size {
	var (
		// This must remain under test for each distinct golang version supported. This unsafe contract could break
		// due to stdlib changes in newer golang versions.
		mapType, hmapInstance = mapTypeAndValue(mapInstance)

		// assume 1 allocated buckets to begin with
		numAllocatedBuckets = 2
	)

	// Best effort escape hatch for unexpected types being passed to this function
	if mapType == nil || hmapInstance == nil {
		return 0
	}

	// the hmap field B represents the log_2 of # of buckets in the hash map
	if hmapInstance.B == 0 {
		numAllocatedBuckets = 1
	}

	// size starts with the number of allocated buckets multiplied by the maptype bucketsize
	size := Size(numAllocatedBuckets<<hmapInstance.B) * Size(mapType.bucketsize)

	// track overflow buckets
	if hmapInstance.noverflow != 0 {
		size += Size(hmapInstance.noverflow) * Size(mapType.bucketsize)
	}

	// track old allocated buckets
	if hmapInstance.oldbuckets != nil {
		size += (2 << (hmapInstance.B - 1)) * Size(mapType.bucketsize)
	}

	// lastly compute the size of the map header
	return Of(*hmapInstance) + size
}

func OfMapValues[K comparable, V any](mapInstance map[K]V) Size {
	size := OfMap(mapInstance)

	for _, value := range mapInstance {
		size += OfContents(value)
	}

	return size
}
