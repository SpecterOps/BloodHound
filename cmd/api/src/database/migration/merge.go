package migration

import (
	"errors"
	"io/fs"
	"sort"
)

type mergedFS struct {
	filesystems []fs.FS
}

func MergedFS(filesystems ...fs.FS) fs.FS {
	return &mergedFS{filesystems: filesystems}
}

func (s *mergedFS) Open(name string) (fs.File, error) {
	for _, filesystem := range s.filesystems {
		file, err := filesystem.Open(name)
		if err == nil {
			return file, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
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

	entriesByName := make(map[string]fs.DirEntry)
	for _, filesystem := range s.filesystems {
		entries, err := fs.ReadDir(filesystem, ".")
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			entriesByName[entry.Name()] = entry
		}

	}
	names := make([]string, 0, len(entriesByName))
	for entryName := range entriesByName {
		names = append(names, entryName)
	}
	sort.Strings(names)
	mergedEntries := make([]fs.DirEntry, 0, len(names))
	for _, entryName := range names {
		mergedEntries = append(mergedEntries, entriesByName[entryName])
	}
	return mergedEntries, nil
}
