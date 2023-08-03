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

package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/specterops/bloodhound/log"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

type Config struct {
	ctx       *cue.Context
	cueConfig *load.Config
}

func (s Config) Value(entrypoint string) (cue.Value, error) {
	if instances := load.Instances([]string{entrypoint}, s.cueConfig); len(instances) == 0 || len(instances) > 1 {
		return cue.Value{}, fmt.Errorf("expected only one instance returned from the library")
	} else if instance := instances[0]; instance.Err != nil {
		return cue.Value{}, instance.Err
	} else {
		// Capture and return the value while also bubbling up the error as part of the return tuple
		value := s.ctx.BuildInstance(instance)
		return value, value.Err()
	}
}

func (s Config) Values(entrypoints ...string) ([]cue.Value, error) {
	instances := load.Instances(entrypoints, s.cueConfig)

	// Fail on the first instance that returns an error
	for _, instance := range instances {
		if instance.Err != nil {
			return nil, instance.Err
		}
	}

	return s.ctx.BuildInstances(instances)
}

type ConfigBuilder struct {
	overlayRootPath string
	overlay         map[string]load.Source
}

func NewConfigBuilder(overlayRootPath string) *ConfigBuilder {
	return &ConfigBuilder{
		overlayRootPath: overlayRootPath,
		overlay:         map[string]load.Source{},
	}
}

func (s *ConfigBuilder) Build() Config {
	return Config{
		ctx: cuecontext.New(),
		cueConfig: &load.Config{
			ModuleRoot: s.overlayRootPath,
			Dir:        s.overlayRootPath,
			Overlay:    s.overlay,
		},
	}
}

func (s *ConfigBuilder) OverlayPath(rootPath string) error {
	return filepath.WalkDir(rootPath, func(path string, dir fs.DirEntry, err error) error {
		if fileInfo, err := os.Lstat(path); err != nil {
			return err
		} else if fileInfo.IsDir() {
			return nil
		}

		if content, err := os.ReadFile(path); err != nil {
			return err
		} else {
			overlayPath := filepath.Join(s.overlayRootPath, strings.TrimPrefix(path, rootPath))

			log.Debugf("Overlaying file: %s to %s", path, overlayPath)
			s.overlay[overlayPath] = load.FromBytes(content)
		}

		return nil
	})
}
