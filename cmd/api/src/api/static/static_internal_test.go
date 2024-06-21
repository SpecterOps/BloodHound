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
	"github.com/specterops/bloodhound/src/api"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
		mockAssetCfg        = AssetConfig{
			FS:         mockFS,
			BasePath:   assetBasePath,
			IndexPath:  indexAssetPath,
			PrefixPath: api.UserInterfacePath,
		}
		handler = MakeAssetHandler(mockAssetCfg)
	)

	t.Run("Test case for an asset that exists", func(t *testing.T) {
		assetsTestMockFile.EXPECT().Read(gomock.AssignableToTypeOf([]byte{})).DoAndReturn(func(target []byte) (int, error) {
			for idx, b := range []byte(expectedOutput) {
				target[idx] = b
			}

			return len(expectedOutput), io.EOF
		})
		mockFS.EXPECT().Open(gomock.Eq("assets/test")).Return(assetsTestMockFile, nil)
		assetsTestMockFile.EXPECT().Close()

		if req, err := http.NewRequest("GET", "", nil); err != nil {
			t.Fatalf("Failed to create request: %v", err)
		} else {
			response := httptest.NewRecorder()

			req.RequestURI = "/ui/test"

			handler.ServeHTTP(response, req)
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, response.Body.String(), expectedOutput)
		}
	})

	t.Run("Mocks for index fallback in asset does not exist cases", func(t *testing.T) {
		assetsIndexMockFile.EXPECT().Read(gomock.AssignableToTypeOf([]byte{})).DoAndReturn(func(target []byte) (int, error) {
			for idx, b := range []byte(expectedOutput) {
				target[idx] = b
			}

			return len(expectedOutput), io.EOF
		}).Times(2)
		mockFS.EXPECT().Open(gomock.Eq("assets/index.html")).Return(assetsIndexMockFile, nil).Times(2)
		assetsIndexMockFile.EXPECT().Close().Times(2)

		// Test case for an asset that does not exist
		mockFS.EXPECT().Open(gomock.Eq("assets/missing")).Return(nil, os.ErrNotExist)

		if req, err := http.NewRequest("GET", "", nil); err != nil {
			t.Fatalf("Failed to create request: %v", err)
		} else {
			response := httptest.NewRecorder()

			req.RequestURI = "/ui/missing"

			handler.ServeHTTP(response, req)
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, response.Body.String(), expectedOutput)
		}
	})

	t.Run("Test case for an asset that does not exist", func(t *testing.T) {
		if req, err := http.NewRequest("GET", "", nil); err != nil {
			t.Fatalf("Failed to create request: %v", err)
		} else {
			response := httptest.NewRecorder()

			req.RequestURI = "/ui/"

			handler.ServeHTTP(response, req)
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, response.Body.String(), expectedOutput)
		}
	})
}
