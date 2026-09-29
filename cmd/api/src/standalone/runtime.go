// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

package standalone

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const runtimeDirectoryName = "runtime/postgres"

func EnsureEmbeddedPostgresRuntime(dataDirectory string) (string, error) {
	runtimeDirectory := filepath.Join(dataDirectory, runtimeDirectoryName)
	pgControlPath := filepath.Join(runtimeDirectory, "bin", "pg_ctl")
	if _, err := os.Stat(pgControlPath); err == nil {
		return runtimeDirectory, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("checking embedded PostgreSQL runtime: %w", err)
	}

	archive, err := embeddedPostgresArchive()
	if err != nil {
		return "", err
	}

	stagingDirectory := runtimeDirectory + ".staging"
	if err := os.RemoveAll(stagingDirectory); err != nil {
		return "", fmt.Errorf("removing incomplete embedded PostgreSQL runtime: %w", err)
	}
	if err := extractArchive(archive, stagingDirectory); err != nil {
		return "", err
	}
	if err := os.Rename(stagingDirectory, runtimeDirectory); err != nil {
		return "", fmt.Errorf("installing embedded PostgreSQL runtime: %w", err)
	}

	if _, err := os.Stat(pgControlPath); err != nil {
		return "", fmt.Errorf("embedded PostgreSQL runtime is missing pg_ctl: %w", err)
	}

	return runtimeDirectory, nil
}

func extractArchive(archive []byte, destination string) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return fmt.Errorf("opening embedded PostgreSQL runtime: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading embedded PostgreSQL runtime: %w", err)
		}
		if header.Typeflag != tar.TypeDir && header.Typeflag != tar.TypeLink && header.Typeflag != tar.TypeReg {
			return fmt.Errorf("embedded PostgreSQL runtime contains unsupported entry %q", header.Name)
		}

		filePath, err := safeArchivePath(destination, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filePath, header.FileInfo().Mode()); err != nil {
				return fmt.Errorf("creating embedded PostgreSQL directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
				return fmt.Errorf("creating embedded PostgreSQL parent directory: %w", err)
			}
			file, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("creating embedded PostgreSQL file: %w", err)
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return fmt.Errorf("extracting embedded PostgreSQL file: %w", err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("closing embedded PostgreSQL file: %w", err)
			}
		case tar.TypeLink:
			linkTargetPath, err := safeArchivePath(destination, header.Linkname)
			if err != nil {
				return err
			}
			if err := os.Link(linkTargetPath, filePath); err != nil {
				return fmt.Errorf("creating embedded PostgreSQL hard link: %w", err)
			}
		}
	}
}

func safeArchivePath(destination string, archivePath string) (string, error) {
	cleanArchivePath := filepath.Clean(archivePath)
	if cleanArchivePath == "." || filepath.IsAbs(cleanArchivePath) || strings.HasPrefix(cleanArchivePath, ".."+string(filepath.Separator)) || cleanArchivePath == ".." {
		return "", fmt.Errorf("embedded PostgreSQL runtime contains unsafe path %q", archivePath)
	}

	return filepath.Join(destination, cleanArchivePath), nil
}
