// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vulndb

import (
	"fmt"
	"github.com/specterops/bloodhound/src/cmd/vizslas/golang/vulndb/version"
)

// New creates a new database from the given entries.
// Errors if there are multiple entries with the same ID.
func New(entries ...Entry) (*Database, error) {
	db := &Database{
		DB:      DBMeta{},
		Modules: make(ModulesIndex),
		Vulns:   make(VulnsIndex),
		Entries: make([]Entry, 0, len(entries)),
	}
	for _, entry := range entries {
		if err := db.Add(entry); err != nil {
			return nil, err
		}
	}
	return db, nil
}

// Add adds new entries to a database, erroring if any of the entries
// is already in the database.
func (db *Database) Add(entries ...Entry) error {
	for _, entry := range entries {
		if err := db.Vulns.add(entry); err != nil {
			return err
		}
		// Only add the entry once we are sure it won't
		// cause an error.
		db.Entries = append(db.Entries, entry)
		db.Modules.add(entry)
		db.DB.add(entry)
	}
	return nil
}

func (dbi *DBMeta) add(entry Entry) {
	if entry.Modified.After(dbi.Modified.Time) {
		dbi.Modified = entry.Modified
	}
}

func (m *ModulesIndex) add(entry Entry) {
	for _, affected := range entry.Affected {
		modulePath := affected.Module.Path
		if _, ok := (*m)[modulePath]; !ok {
			(*m)[modulePath] = &IndexedModule{
				Path:  modulePath,
				Vulns: []ModuleVuln{},
			}
		}
		module := (*m)[modulePath]
		module.Vulns = append(module.Vulns, ModuleVuln{
			ID:       entry.ID,
			Modified: entry.Modified,
			Fixed:    latestFixedVersion(affected.Ranges),
		})
	}
}

func (v *VulnsIndex) add(entry Entry) error {
	if _, ok := (*v)[entry.ID]; ok {
		return fmt.Errorf("id %q appears twice in database", entry.ID)
	}
	(*v)[entry.ID] = &Vuln{
		ID:       entry.ID,
		Modified: entry.Modified,
		Aliases:  entry.Aliases,
	}
	return nil
}

func latestFixedVersion(ranges []Range) string {
	var latestFixed string
	for _, r := range ranges {
		if r.Type == RangeTypeSemver {
			for _, e := range r.Events {
				if fixed := e.Fixed; fixed != "" && version.Before(latestFixed, fixed) {
					latestFixed = fixed
				}
			}
			// If the vulnerability was re-introduced after the latest fix
			// we found, there is no latest fix for this range.
			for _, e := range r.Events {
				if introduced := e.Introduced; introduced != "" && introduced != "0" && version.Before(latestFixed, introduced) {
					latestFixed = ""
					break
				}
			}
		}
	}
	return string(latestFixed)
}
