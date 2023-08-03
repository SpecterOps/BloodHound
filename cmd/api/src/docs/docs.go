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

package docs

import (
	"bytes"
	"embed"
	"encoding/json"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"

	"github.com/alecthomas/template"
)

//go:embed json/definitions/* json/paths/*
var jsonFiles embed.FS

func NewSwaggerConfig(configFns ...func(config *httpSwagger.Config)) *httpSwagger.Config {
	config := httpSwagger.Config{
		URL:                  "doc.json",
		DocExpansion:         "list",
		DomID:                "swagger-ui",
		InstanceName:         "swagger",
		DeepLinking:          true,
		PersistAuthorization: false,
	}

	for _, fn := range configFns {
		fn(&config)
	}

	if config.InstanceName == "" {
		config.InstanceName = swag.Name
	}

	return &config
}

type SwaggerInfo struct {
	Version      string
	Host         string
	BasePath     string
	Schemes      []string
	Title        string
	Description  string
	Paths        string
	Definitions  string
	SupportName  string
	SupportUrl   string
	SupportEmail string
}

const (
	v2 = "v2"

	doc = `
	{
	  "schemes": {{ marshal .Schemes }},
	  "openapi": "3.0.3",
	  "info": {
		"description": "{{.Description}}",
		"title": "{{.Title}}",
		"termsOfService": "TBD",
		"contact": {
		  "name": "{{.SupportName}}",
		  "url": "{{.SupportUrl}}",
		  "email": "{{.SupportEmail}}"
		},
		"license": {
		  "name": "Apache-2.0",
		  "url": "http://www.apache.org/licenses/LICENSE-2.0"
		},
		"version": "{{.Version}}"
	  },
	  "basePath": "{{.BasePath}}",
	  "paths": {{.Paths}},
	  "definitions": {{.Definitions}},
	  "consumes": [
		"application/json"
	  ],
	  "produces": [
		"application/json"
	  ],
	  "securitySchemes": {
		"OAuth2PasswordBearer": {
		  "type": "oauth2",
		  "flow": "password",
		  "tokenUrl": "https://localhost:8080/api/v2/login"
		}
	  },
	  "components": {
		"responses": {
			"defaultError": {
				"description": "Standard response for any errors that may occur.",
				"content": {
					"application/json": {
						"schema": { "$ref": "#/definitions/api.ErrorWrapper" }
					}
				}
			}
		}
	  }
	}
	`
)

func (s *SwaggerInfo) readDefinitions(version string) error {
	// Combine all our definitions JSON files into a single JSON Object
	if definitionsFiles, err := jsonFiles.ReadDir("json/definitions"); err != nil {
		return err
	} else {
		var definitionMap = make(map[string]any)

		for _, file := range definitionsFiles {
			var definition map[string]any

			if definitionJson, err := jsonFiles.ReadFile("json/definitions/" + file.Name()); err != nil {
				return err
			} else if err := json.Unmarshal(definitionJson, &definition); err != nil {
				return err
			} else {
				for key, val := range definition {
					definitionMap[key] = val
				}
			}
		}

		if definitionBytes, err := json.Marshal(definitionMap); err != nil {
			return err
		} else {
			s.Definitions = string(definitionBytes)
		}
	}
	return nil
}

func (s *SwaggerInfo) readPaths(location string) error {
	// Combine all our paths JSON files into a single JSON Object
	if pathsFiles, err := jsonFiles.ReadDir(location); err != nil {
		return err
	} else {
		var pathMap = make(map[string]any)

		for _, file := range pathsFiles {
			var path map[string]any

			if pathJson, err := jsonFiles.ReadFile(location + "/" + file.Name()); err != nil {
				return err
			} else if err := json.Unmarshal(pathJson, &path); err != nil {
				return err
			} else {
				for key, val := range path {
					pathMap[key] = val
				}
			}
		}

		if pathBytes, err := json.Marshal(pathMap); err != nil {
			return err
		} else {
			s.Paths = string(pathBytes)
		}
	}
	return nil
}

type SwaggerV2 struct{}

func (s *SwaggerV2) ReadDoc(sInfo SwaggerInfo) string {
	if err := sInfo.readDefinitions(v2); err != nil {
		return doc
	} else if err := sInfo.readPaths("json/paths/v2"); err != nil {
		return doc
	} else {
		t, err := template.New("swagger_info").Funcs(template.FuncMap{
			"marshal": func(v any) string {
				a, _ := json.Marshal(v)
				return string(a)
			},
		}).Parse(doc)
		if err != nil {
			return doc
		}

		var tpl bytes.Buffer
		if err := t.Execute(&tpl, sInfo); err != nil {
			return doc
		}

		return tpl.String()
	}
}

const SwaggerIndexTemplate = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta http-equiv="refresh" content="5;url=/ui/api-explorer" />
  <title>Swagger UI</title>
</head>
<body>
	<h1>The API documentation has moved and requires an authenticated user.</h1>
	<p>Redirecting in <span id="counter">5</span> seconds...</p>
	<script>
	const counter = document.querySelector('#counter');
	const interval = setInterval(function(){
		counter.textContent = counter.textContent * 1 - 1;
		if (counter.textContent == 0) {
			clearInterval(interval);
		}
	}, 1000);
	</script>
</body>
</html>
`
