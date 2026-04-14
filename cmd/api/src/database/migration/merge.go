package migration

import (
	"io/fs"
	"path"
	"sort"
	"strings"
)

// goose flow :

// Goose calls ReadDir(".") on mergedFS based off of the constructor
// ReadDir returns DirEntry with sorted values ["001_init.sql", "002_init.sql", "005_users.sql", "v9/007_colapse.sql"]
// Goose iterates that list and calls Open("001_init.sql") which Open() returns the fs.Fs and allows goose to
// 		view contents of the file and execute sql depending on up or down --prefixes

// Goose will iterate and compare between the goose_db_version table from the PG database if a migration has been run
// and will run missing and out of order migrations

type mergedFS struct {
	filesystems []fs.FS
	pathIndex   map[string]indexedPath
}

type indexedPath struct {
	filesystem fs.FS
	fullPath   string
	dirEntry   fs.DirEntry
}

func MergedFS(filesystems ...fs.FS) fs.FS {
	return &mergedFS{
		filesystems: filesystems,
		pathIndex:   nil,
	}
}

func (s *mergedFS) buildIndex() error {
	s.pathIndex = make(map[string]indexedPath)
	for _, filesystem := range s.filesystems {
		// filesystem -> migrations folder specified from the constructor in migration.go
		// dirPath "." -> start search from the root, in this case we treat the "migrations" folderas the root
		if err := s.collectEntries(filesystem, "."); err != nil {
			return err
		}
	}
	return nil
}

// collectEntries recursively walks dirPath within a single filesystem,
// adding each .sql file to the shared pathIndex.
func (s *mergedFS) collectEntries(filesystem fs.FS, dirPath string) error {
	entries, err := fs.ReadDir(filesystem, dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := path.Join(dirPath, entry.Name())
		if entry.IsDir() {
			if err := s.collectEntries(filesystem, entryPath); err != nil {
				return err
			}
		} else if strings.HasSuffix(entry.Name(), ".sql") {
			// Only index if not already present — first filesystem wins.
			if _, exists := s.pathIndex[entry.Name()]; !exists {
				s.pathIndex[entry.Name()] = indexedPath{
					filesystem: filesystem, // {address to bhe or foss file system} // used for s.Open func to determine which folder path to search from
					fullPath:   entryPath,  // {"timestamp_something.sql"} full relative path for goose
					dirEntry:   entry,      // {address to directory} necessary for goose
				}
			}
		}
	}
	return nil
}

// Open is run by goose to "Open" a fs based on the string name it is looking for to execute the migrations of that
// .sql file. We return a fs.File to pass back to goose so that the sql can be executed.
func (s *mergedFS) Open(name string) (fs.File, error) {
	// Ensure the index is built before any lookups.
	if s.pathIndex == nil {
		if err := s.buildIndex(); err != nil {
			return nil, err
		}
	}

	// Single map lookup replaces both the flat scan and recursive fallback.
	// If the key of the file it is looking for 00001_init.sql exists then open and return to goose to execute sql
	// from that file
	if indexed, exists := s.pathIndex[name]; exists {
		return indexed.filesystem.Open(indexed.fullPath)
	}

	return nil, &fs.PathError{
		Op:   "open",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

// ReadDir returns a sorted slice of all migration file entries
// from foss and bhe for the goose provider.
func (s *mergedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name != "." {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: name,
			Err:  fs.ErrNotExist,
		}
	}

	// Ensure the index is built before reading.
	if s.pathIndex == nil {
		if err := s.buildIndex(); err != nil {
			return nil, err
		}
	}

	// Collect and sort names from the shared index.
	names := make([]string, 0, len(s.pathIndex))
	for entryName := range s.pathIndex {
		names = append(names, entryName)
	}
	sort.Strings(names)

	mergedEntries := make([]fs.DirEntry, 0, len(names))
	for _, entryName := range names {
		mergedEntries = append(mergedEntries, s.pathIndex[entryName].dirEntry)
	}
	return mergedEntries, nil
}
