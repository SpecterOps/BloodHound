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
	"os"
	"path/filepath"
	"testing"

	"github.com/specterops/dawgs/drivers/pg"
	"github.com/stretchr/testify/require"
)

func TestPreparePersistsStandaloneConfiguration(t *testing.T) {
	dataDirectory := t.TempDir()

	instance, err := Prepare(Options{
		DataDirectory:        dataDirectory,
		PostgresBinDirectory: "/test/postgres/bin",
		APIPort:              8080,
	})
	require.NoError(t, err)
	require.Equal(t, pg.DriverName, instance.Configuration.GraphDriver)
	require.Equal(t, "127.0.0.1:8080", instance.Configuration.BindAddress)
	require.Equal(t, "bloodhound", instance.Configuration.Database.Database)
	require.Equal(t, "bloodhound", instance.Configuration.Database.Username)
	require.Equal(t, filepath.Join(dataDirectory, "collectors"), instance.Configuration.CollectorsBasePath)
	require.Contains(t, instance.Configuration.Database.Connection, "sslmode=disable")
	require.Equal(t, "/test/postgres/bin", instance.Postgres.binDirectory)
	require.NotEmpty(t, instance.Configuration.Crypto.JWT.SigningKey)
	require.NotEmpty(t, instance.Configuration.Database.Secret)

	configurationPath := filepath.Join(dataDirectory, configurationFileName)
	fileInfo, err := os.Stat(configurationPath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())

	restartedInstance, err := Prepare(Options{
		DataDirectory:        dataDirectory,
		PostgresBinDirectory: "/test/postgres/bin",
		APIPort:              8080,
	})
	require.NoError(t, err)
	require.Equal(t, instance.Configuration.Crypto.JWT.SigningKey, restartedInstance.Configuration.Crypto.JWT.SigningKey)
	require.Equal(t, instance.Configuration.Database.Secret, restartedInstance.Configuration.Database.Secret)
}

func TestExtractArchive(t *testing.T) {
	var archive bytes.Buffer
	gzipWriter := gzip.NewWriter(&archive)
	tarWriter := tar.NewWriter(gzipWriter)
	require.NoError(t, tarWriter.WriteHeader(&tar.Header{Name: "bin", Typeflag: tar.TypeDir, Mode: 0o755}))
	require.NoError(t, tarWriter.WriteHeader(&tar.Header{Name: "bin/pg_ctl", Mode: 0o755, Size: int64(len("binary"))}))
	_, err := tarWriter.Write([]byte("binary"))
	require.NoError(t, err)
	require.NoError(t, tarWriter.WriteHeader(&tar.Header{Name: "bin/postgres", Typeflag: tar.TypeLink, Linkname: "bin/pg_ctl"}))
	require.NoError(t, tarWriter.Close())
	require.NoError(t, gzipWriter.Close())

	destination := t.TempDir()
	require.NoError(t, extractArchive(archive.Bytes(), destination))

	contents, err := os.ReadFile(filepath.Join(destination, "bin", "pg_ctl"))
	require.NoError(t, err)
	require.Equal(t, "binary", string(contents))

	linkedContents, err := os.ReadFile(filepath.Join(destination, "bin", "postgres"))
	require.NoError(t, err)
	require.Equal(t, "binary", string(linkedContents))
}

func TestSafeArchivePathRejectsTraversal(t *testing.T) {
	_, err := safeArchivePath(t.TempDir(), "../outside")
	require.Error(t, err)
}
