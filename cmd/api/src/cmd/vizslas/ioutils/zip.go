package ioutils

import (
	"archive/zip"
	"encoding/json"
	"io"
	"io/fs"
	"strings"
)

type FileFilter struct {
	MustContainOneOf []string
	AcceptsExtension func(filePath string, fileInfo fs.FileInfo) bool
}

func (s FileFilter) Accepts(filePath string, fileInfo fs.FileInfo) bool {
	var (
		accepted = true
		and      = func(value bool) {
			accepted = accepted && value
		}
	)

	if s.AcceptsExtension != nil {
		and(s.AcceptsExtension(filePath, fileInfo))
	}

	and(func() bool {
		if len(s.MustContainOneOf) == 0 {
			return true
		}

		for _, mustContainOneOfEntry := range s.MustContainOneOf {
			if strings.Contains(filePath, mustContainOneOfEntry) {
				return true
			}
		}

		return false
	}())

	return accepted
}

func JSONDecodeReader[T any](reader io.Reader) (T, error) {
	var value T
	return value, json.NewDecoder(reader).Decode(&value)
}

func JSONDecodeZipFile[T any](zipReader *zip.Reader, archivePath string) (T, error) {
	if zipFileIn, err := zipReader.Open(archivePath); err != nil {
		var emptyT T
		return emptyT, err
	} else {
		defer zipFileIn.Close()
		return JSONDecodeReader[T](zipFileIn)
	}
}
