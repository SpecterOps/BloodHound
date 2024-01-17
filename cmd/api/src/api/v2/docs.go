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

package v2

import (
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"sync"

	swaggerFiles "github.com/swaggo/files"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/net/webdav"

	"github.com/specterops/bloodhound/src/docs"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
)

var (
	sInfo = docs.SwaggerInfo{
		Version:      "0.2",
		Host:         "localhost:8080",
		BasePath:     "/",
		Schemes:      []string{},
		Title:        "Bloodhound Community Edition API",
		Description:  "This is the API for all your BHCE needs",
		SupportName:  "BHCE Support",
		SupportUrl:   "https://bloodhoundgang.herokuapp.com/",
		SupportEmail: "bloodhoundenterprise@specterops.io",
	}
)

func SwaggerHandler(configFns ...func(config *httpSwagger.Config)) http.HandlerFunc {
	var (
		once    sync.Once
		Handler = &webdav.Handler{
			FileSystem: swaggerFiles.FS,
			LockSystem: webdav.NewMemLS(),
		}
	)

	config := docs.NewSwaggerConfig(configFns...)

	// create a template with name
	index, _ := template.New("v2_swagger_index.html").Parse(docs.SwaggerIndexTemplate)

	re := regexp.MustCompile(`^(.*/)([^?].*)?[?|.]*$`)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

			return
		}

		matches := re.FindStringSubmatch(r.RequestURI)

		path := matches[2]

		handler := Handler
		once.Do(func() {
			handler.Prefix = matches[1]
		})

		switch filepath.Ext(path) {
		case ".html":
			w.Header().Set(headers.ContentType.String(), mediatypes.TextHtml.WithCharset("utf-8"))
		case ".css":
			w.Header().Set(headers.ContentType.String(), mediatypes.TextCss.WithCharset("utf-8"))
		case ".js":
			w.Header().Set(headers.ContentType.String(), mediatypes.TextJavascript.String())
		case ".png":
			w.Header().Set(headers.ContentType.String(), mediatypes.ImagePng.String())
		case ".json":
			w.Header().Set(headers.ContentType.String(), mediatypes.ApplicationJson.WithCharset("utf-8"))
		}

		swagger := docs.SwaggerV2{}
		switch path {
		case "index.html":
			_ = index.Execute(w, config)
		case "doc.json":
			doc := swagger.ReadDoc(sInfo)
			_, _ = w.Write([]byte(doc))
		case "":
			http.Redirect(w, r, handler.Prefix+"index.html", http.StatusMovedPermanently)
		default:
			handler.ServeHTTP(w, r)
		}
	}
}
