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

package static

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/vendormocks/io/fs"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBHCEStaticHandler(t *testing.T) {
	const expectedOutput = "test"

	var (
		mockController      = gomock.NewController(t)
		assetsTestMockFile  = fs.NewMockFile(mockController)
		assetsIndexMockFile = fs.NewMockFile(mockController)
		mockFS              = fs.NewMockFS(mockController)
		mockFileInfo        = fs.NewMockFileInfo(mockController)
		mockAssetCfg        = AssetConfig{
			FS:         mockFS,
			BasePath:   assetBasePath,
			IndexPath:  indexAssetPath,
			PrefixPath: api.UserInterfacePath,
		}
		handler = MakeAssetHandler(mockAssetCfg)
	)

	t.Run("success - asset exists", func(t *testing.T) {

		mockFS.EXPECT().Open(gomock.Eq("assets/test.html")).Return(assetsTestMockFile, nil)
		assetsTestMockFile.EXPECT().Stat().Return(mockFileInfo, nil)
		mockFileInfo.EXPECT().Name().Return("test.html")
		assetsTestMockFile.EXPECT().Read(gomock.AssignableToTypeOf([]byte{})).DoAndReturn(func(target []byte) (int, error) {
			for idx, b := range []byte(expectedOutput) {
				target[idx] = b
			}

			return len(expectedOutput), io.EOF
		})
		assetsTestMockFile.EXPECT().Close()

		if req, err := http.NewRequest("GET", "", nil); err != nil {
			t.Fatalf("Failed to create request: %v", err)
		} else {
			response := httptest.NewRecorder()

			req.RequestURI = "/test.html"

			handler.ServeHTTP(response, req)
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, expectedOutput, response.Body.String())
			require.Equal(t, "text/html; charset=utf-8", response.Header().Get("Content-Type"))
		}
	})

	t.Run("success - fallback to index on asset that does not exist", func(t *testing.T) {
		// Test case for an asset that does not exist
		mockFS.EXPECT().Open(gomock.Eq("assets/missing.pdf")).Return(nil, os.ErrNotExist)
		mockFS.EXPECT().Open(gomock.Eq("assets/index.html")).Return(assetsIndexMockFile, nil).Times(1)
		assetsIndexMockFile.EXPECT().Stat().Return(mockFileInfo, nil)
		mockFileInfo.EXPECT().Name().Return("test.html")
		assetsIndexMockFile.EXPECT().Read(gomock.AssignableToTypeOf([]byte{})).DoAndReturn(func(target []byte) (int, error) {
			for idx, b := range []byte(expectedOutput) {
				target[idx] = b
			}

			return len(expectedOutput), io.EOF
		}).Times(1)
		assetsIndexMockFile.EXPECT().Close().Times(1)

		if req, err := http.NewRequest("GET", "", nil); err != nil {
			t.Fatalf("Failed to create request: %v", err)
		} else {
			response := httptest.NewRecorder()

			req.RequestURI = "/ui/missing.pdf"

			handler.ServeHTTP(response, req)
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, expectedOutput, response.Body.String())
			require.Equal(t, "text/html; charset=utf-8", response.Header().Get("Content-Type"))
		}
	})

	t.Run("fail - error loading index.html", func(t *testing.T) {
		mockFS.EXPECT().Open(gomock.Eq("assets/missing.pdf")).Return(nil, os.ErrNotExist)
		mockFS.EXPECT().Open(gomock.Eq("assets/index.html")).Return(nil, os.ErrNotExist)

		if req, err := http.NewRequest("GET", "", nil); err != nil {
			t.Fatalf("Failed to create request: %v", err)
		} else {
			response := httptest.NewRecorder()

			req.RequestURI = "missing.pdf"

			handler.ServeHTTP(response, req)
			require.Equal(t, http.StatusNotFound, response.Code)
			require.Contains(t, response.Body.String(), "resource not found")
		}
	})
	t.Run("fail - error retrieving file stats", func(t *testing.T) {
		mockFS.EXPECT().Open(gomock.Eq("assets/index.html")).Return(assetsIndexMockFile, nil)
		assetsIndexMockFile.EXPECT().Stat().Return(nil, fmt.Errorf("error retrieving file stats"))

		if req, err := http.NewRequest("GET", "", nil); err != nil {
			t.Fatalf("Failed to create request: %v", err)
		} else {
			response := httptest.NewRecorder()

			req.RequestURI = "index.html"

			handler.ServeHTTP(response, req)
			require.Equal(t, http.StatusInternalServerError, response.Code)
			require.Contains(t, response.Body.String(), api.ErrorResponseDetailsInternalServerError)
		}
	})
}
