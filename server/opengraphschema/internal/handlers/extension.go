// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
)

type ExtensionMinimalView struct {
	ExtensionID int32  `json:"extension_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Namespace   string `json:"namespace"`
	Version     string `json:"version"`
}

type ExtensionView struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Version   string    `json:"version"`
	IsBuiltIn bool      `json:"is_builtin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ExtensionsView struct {
	Extensions []ExtensionView `json:"extensions"`
}

func (s ExtensionsView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

// Note: Display name is used for name for extension view
func buildExtensionView(extension services.Extension) ExtensionView {
	return ExtensionView{
		ID:        extension.ID,
		Name:      extension.DisplayName,
		Namespace: extension.Namespace,
		Version:   extension.Version,
		IsBuiltIn: extension.IsBuiltin,
		CreatedAt: extension.CreatedAt,
		UpdatedAt: extension.UpdatedAt,
	}
}

func buildExtensionsView(extensions []services.Extension) ExtensionsView {
	var view = ExtensionsView{Extensions: make([]ExtensionView, 0, len(extensions))}

	for _, extension := range extensions {
		view.Extensions = append(view.Extensions, buildExtensionView(extension))
	}

	return view
}

func (s *Handlers) ListExtensions(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()
	if extensions, err := s.service.ListExtensions(ctx); err != nil {
		responses.WriteInternalServerError(ctx, err, response)
	} else {
		responses.WriteBasic(ctx, buildExtensionsView(extensions), http.StatusOK, response)
	}
}
