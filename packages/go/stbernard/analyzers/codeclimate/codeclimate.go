// Copyright 2024 Specter Ops, Inc.
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

package codeclimate

import (
	"fmt"
	"maps"
	"slices"
)

// Entry represents a simple CodeClimate-like entry
type Entry struct {
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	Location    Location `json:"location"`
}

// Location represents a code path and lines for an entry
type Location struct {
	Path  string `json:"path"`
	Lines Lines  `json:"lines"`
}

// Lines represents only the beginning line for the location
type Lines struct {
	Begin uint64 `json:"begin"`
}

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityMinor    Severity = "minor"
	SeverityMajor    Severity = "major"
	SeverityCritical Severity = "critical"
	SeverityBlocker  Severity = "blocker"
	SeverityInvalid  Severity = "invalid"
)

func ParseSeverity(severityStr string) (Severity, error) {
	switch severity := Severity(severityStr); severity {
	case SeverityBlocker, SeverityCritical, SeverityInfo, SeverityMinor, SeverityMajor:
		return severity, nil
	}

	return SeverityInvalid, fmt.Errorf("invalid severity: %s", severityStr)
}

func (s Severity) Priority() int {
	switch s {
	case SeverityInfo:
		return 1
	case SeverityMinor:
		return 2
	case SeverityMajor:
		return 3
	case SeverityCritical:
		return 4
	case SeverityBlocker:
		return 5
	default:
		return -1
	}
}

// EntryMap represents a grouping of code climate entries grouped by file path
type EntryMap map[string][]Entry

func (s EntryMap) Merge(other EntryMap) {
	for otherKey, otherValue := range other {
		s[otherKey] = append(s[otherKey], otherValue...)
	}
}

func (s EntryMap) Copy() EntryMap {
	copied := make(EntryMap, len(s))

	for key, value := range s {
		copied[key] = append(copied[key], value...)
	}

	return copied
}

// SeverityMap represents a grouping of EntryMap instances grouped by severity
type SeverityMap map[Severity]EntryMap

func CombineSeverityMaps(maps ...SeverityMap) SeverityMap {
	if len(maps) == 0 {
		return SeverityMap{}
	}

	combined := maps[0].Copy()

	for _, merge := range maps[1:] {
		combined.Merge(merge)
	}

	return combined
}

func NewSeverityMap(entries []Entry) SeverityMap {
	severityMap := SeverityMap{}

	// Group existing entries by severity and file path
	for _, entry := range entries {
		if existingEntries, hasExisting := severityMap[entry.Severity]; hasExisting {
			existingEntries[entry.Location.Path] = append(existingEntries[entry.Location.Path], entry)
		} else {
			severityMap[entry.Severity] = EntryMap{
				entry.Location.Path: []Entry{
					entry,
				},
			}
		}
	}

	return severityMap
}

func (s SeverityMap) Merge(other SeverityMap) {
	for otherKey, otherValue := range other {
		if existing, hasExisting := s[otherKey]; hasExisting {
			existing.Merge(otherValue)
		} else {
			s[otherKey] = otherValue.Copy()
		}
	}
}

func (s SeverityMap) SortedSeverities() []Severity {
	return slices.SortedFunc(maps.Keys(s), func(a Severity, b Severity) int {
		return a.Priority() - b.Priority()
	})
}

func (s SeverityMap) HasGreaterSeverity(than Severity) bool {
	for severity := range s {
		if severity.Priority() > than.Priority() {
			return true
		}
	}

	return false
}

func (s SeverityMap) Copy() SeverityMap {
	copied := make(SeverityMap, len(s))

	for key, value := range s {
		copied[key] = value.Copy()
	}

	return copied
}
