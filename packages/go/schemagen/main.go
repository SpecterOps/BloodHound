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

package main

import (
	"os"
	"path/filepath"

	"cuelang.org/go/cue/errors"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/schemagen/generator"
	"github.com/specterops/bloodhound/schemagen/model"
	"github.com/specterops/bloodhound/schemagen/tsgen"
)

type Schema struct {
	Common          model.Graph
	Azure           model.Azure
	ActiveDirectory model.ActiveDirectory
}

func GenerateGolang(projectRoot string, rootSchema Schema) error {
	if err := generator.GenerateGolangSchemaTypes("graphschema", filepath.Join(projectRoot, "packages/go/graphschema")); err != nil {
		return err
	}

	if err := generator.GenerateGolangGraphModel("common", filepath.Join(projectRoot, "packages/go/graphschema/common"), rootSchema.Common); err != nil {
		return err
	}

	if err := generator.GenerateGolangActiveDirectory("ad", filepath.Join(projectRoot, "packages/go/graphschema/ad"), rootSchema.ActiveDirectory); err != nil {
		return err
	}

	if err := generator.GenerateGolangAzure("azure", filepath.Join(projectRoot, "packages/go/graphschema/azure"), rootSchema.Azure); err != nil {
		return err
	}

	return nil
}

func GenerateSharedTypeScript(projectRoot string, rootSchema Schema) error {
	root := tsgen.NewFile("graph_schema", filepath.Join(projectRoot, "packages/javascript/bh-shared-ui/src/graphSchema.ts"))

	generator.GenerateTypeScriptActiveDirectory(root, rootSchema.ActiveDirectory)
	generator.GenerateTypeScriptAzure(root, rootSchema.Azure)
	generator.GenerateTypeScriptCommon(root, rootSchema.Common)

	return root.Write(os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
}

func main() {
	log.Configure(log.DefaultConfiguration().WithLevel(log.LevelDebug))

	cfgBuilder := generator.NewConfigBuilder("/schemas")

	if projectRoot, err := generator.FindGolangWorkspaceRoot(); err != nil {
		log.Fatalf("Error finding project root: %v", err)
	} else {
		log.Infof("Project root is %s", projectRoot)

		if err := cfgBuilder.OverlayPath(filepath.Join(projectRoot, "packages/cue")); err != nil {
			log.Fatalf("Error: %v", err)
		}

		cfg := cfgBuilder.Build()

		if bhInstance, err := cfg.Value("/schemas/bh/bh.cue"); err != nil {
			log.Fatalf("Error: %v", errors.Details(err, nil))
		} else {
			var bhModels Schema

			if err := bhInstance.Decode(&bhModels); err != nil {
				log.Fatalf("Error: %v", errors.Details(err, nil))
			}

			if err := GenerateGolang(projectRoot, bhModels); err != nil {
				log.Fatalf("Error %v", err)
			}

			if err := GenerateSharedTypeScript(projectRoot, bhModels); err != nil {
				log.Fatalf("Error %v", err)
			}
		}
	}
}
