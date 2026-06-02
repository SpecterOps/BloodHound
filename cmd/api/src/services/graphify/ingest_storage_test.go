// Copyright 2026 Specter Ops, Inc.
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

package graphify_test

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/stretchr/testify/require"
)

type zipTestFile struct {
	name    string
	content []byte
}

func writeZipFile(t *testing.T, archivePath string, files []zipTestFile) {
	var (
		archiveFile *os.File
		zipWriter   *zip.Writer
		err         error
	)

	archiveFile, err = os.Create(archivePath)
	require.NoError(t, err)
	defer archiveFile.Close()

	zipWriter = zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	for _, file := range files {
		fileWriter, err := zipWriter.Create(file.name)
		require.NoError(t, err)

		_, err = fileWriter.Write(file.content)
		require.NoError(t, err)
	}

	_, err = zipWriter.Create("empty-directory/")
	require.NoError(t, err)
}

func TestExtractIngestFilesReturnsJSONFile(t *testing.T) {
	var (
		ctx              = context.Background()
		configuration    = config.Configuration{WorkDir: t.TempDir()}
		storedFileName   = filepath.Join(configuration.WorkDir, "ingest.json")
		providedFileName = "provided.json"
	)

	fileData, err := graphify.ExtractIngestFiles(ctx, configuration, storedFileName, providedFileName, model.FileTypeJson, "unused")
	require.NoError(t, err)
	require.Equal(t, []graphify.IngestFileData{
		{
			Name:   providedFileName,
			Path:   storedFileName,
			Errors: []string{},
		},
	}, fileData)
}

func TestExtractIngestFilesExpandsZIPIntoTempDirectory(t *testing.T) {
	var (
		ctx              = context.Background()
		configuration    = config.Configuration{WorkDir: t.TempDir()}
		archivePath      = filepath.Join(configuration.WorkDir, "archive.zip")
		providedFileName = "archive.zip"
		tempFilePrefix   = "file_upload_job7_"
		bomFileContent   = append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"meta":{"type":"domains","version":5,"count":0},"data":[]}`)...)
		plainFileContent = []byte(`{"meta":{"type":"users","version":5,"count":0},"data":[]}`)
	)

	require.NoError(t, os.MkdirAll(configuration.TempDirectory(), 0o755))
	writeZipFile(t, archivePath, []zipTestFile{
		{
			name:    "domains.json",
			content: bomFileContent,
		},
		{
			name:    "users.json",
			content: plainFileContent,
		},
	})

	fileData, err := graphify.ExtractIngestFiles(ctx, configuration, archivePath, providedFileName, model.FileTypeZip, tempFilePrefix)
	require.NoError(t, err)
	require.Len(t, fileData, 2)
	require.NoFileExists(t, archivePath)

	for _, file := range fileData {
		require.Equal(t, providedFileName, file.ParentFile)
		require.Empty(t, file.Errors)
		require.DirExists(t, filepath.Dir(file.Path))
		require.FileExists(t, file.Path)
		require.Contains(t, filepath.Base(file.Path), tempFilePrefix)
	}

	domainsContent, err := os.ReadFile(fileData[0].Path)
	require.NoError(t, err)
	require.Equal(t, []byte(`{"meta":{"type":"domains","version":5,"count":0},"data":[]}`), domainsContent)

	usersContent, err := os.ReadFile(fileData[1].Path)
	require.NoError(t, err)
	require.Equal(t, plainFileContent, usersContent)
}
