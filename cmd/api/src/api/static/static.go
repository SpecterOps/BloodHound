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
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/specterops/bloodhound/src/utils"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
)

type AssetConfig struct {
	FS         fs.FS
	PrefixPath string
	BasePath   string
	IndexPath  string
}

func MakeAssetHandler(cfg AssetConfig) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		serve(cfg, response, request)
	})
}

// fetchAsset will attempt to find a static asset at the given path. If the asset does not exist, fetchAsset will instead
// return the index.html asset. This is done because the UI will change with the browser URI depending on the UI view.
// This results in the browser asking for assets that do not exist upon browser refresh.
func fetchAsset(cfg AssetConfig, assetPath string) (io.ReadCloser, error) {
	if fin, err := cfg.FS.Open(filepath.Join(cfg.BasePath, assetPath)); err != nil {
		if cfg.IndexPath != "" {
			return cfg.FS.Open(filepath.Join(cfg.BasePath, cfg.IndexPath))
		}
		return nil, err
	} else {
		return fin, nil
	}
}

func serve(cfg AssetConfig, response http.ResponseWriter, request *http.Request) {
	var (
		// Strip off the ui path prefix from the request URI path
		assetPath = strings.TrimPrefix(request.RequestURI, cfg.PrefixPath)
	)

	// Rewrite references to root as "index.html" - without this, the embed.FS will happily return a "directory" FD
	// instead of failing to open the path
	if assetPath == "" || assetPath == "/" {
		assetPath = cfg.IndexPath
	}

	if fin, err := fetchAsset(cfg, assetPath); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else {
		defer fin.Close()

		var (
			assetExtension = filepath.Ext(assetPath)
			contentType    = mime.TypeByExtension(assetExtension)
		)

		// default to "text/html; charset=utf-8" if there was no file extension detected
		if contentType == "" {
			contentType = mime.TypeByExtension(".html")
		}

		response.Header().Set(headers.ContentType.String(), contentType)
		response.Header().Set(headers.StrictTransportSecurity.String(), utils.HSTSSetting)

		if _, err := io.Copy(response, fin); err != nil {
			log.Errorf("Failed flushing static file content for asset %s to client: %v", assetPath, err)
		}
	}
}
