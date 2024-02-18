package cve

import (
	"archive/zip"
	"context"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/cmd/vizslas/ioutils"
	"os"
	"sync"
)

type Database struct {
	lock    *sync.Mutex
	entries map[string]*Entry
}

func NewDatabase() *Database {
	return &Database{
		lock:    &sync.Mutex{},
		entries: map[string]*Entry{},
	}
}

func (s *Database) AddEntry(entry *Entry) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.entries[entry.Metadata.ID] = entry
}

func (s *Database) LookupCVE(metadataID string) (*Entry, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	entry, hasEntry := s.entries[metadataID]
	return entry, hasEntry
}

func ReadArchive(ctx context.Context, archivePath string, filter ioutils.FileFilter) (*Database, error) {
	if fileIn, err := os.Open(archivePath); err != nil {
		return nil, err
	} else {
		defer fileIn.Close()

		if zipFileInfo, err := fileIn.Stat(); err != nil {
			return nil, err
		} else if zipReader, err := zip.NewReader(fileIn, zipFileInfo.Size()); err != nil {
			return nil, err
		} else {
			database := NewDatabase()

			for _, file := range zipReader.File {
				// Check the context for early exit
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}

				// Check to see if this file is accepted
				if !filter.Accepts(file.Name, file.FileInfo()) {
					continue
				}

				// Read and marshall the CVE JSON file
				if entry, err := ioutils.JSONDecodeZipFile[*Entry](zipReader, file.Name); err != nil {
					log.Errorf("Zip file reader error: %v", err)
				} else {
					database.AddEntry(entry)
				}
			}

			log.Infof("Loaded %d entries", len(database.entries))

			return database, nil
		}
	}
}
