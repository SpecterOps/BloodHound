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
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/specterops/bloodhound/conftool/config"
)

func main() {
	var (
		path       string
		tuneMillis int
		skipArgon2 bool
	)
	flag.StringVar(&path, "file", "/tmp/bloodhound.config.json", "path to create config file")
	flag.IntVar(&tuneMillis, "tuneMillis", 500, "number of milliseconds to tune to")
	flag.BoolVar(&skipArgon2, "skipArgon2", false, "skip argon2")
	flag.Parse()

	if configfile, err := os.Create(path); err != nil {
		log.Fatalf("Could not create config file %s: %v", path, err)
	} else {
		defer configfile.Close()

		if !skipArgon2 {
			log.Printf("Tuning Argon2 parameters to target %d milliseconds. This might take some time.", tuneMillis)
		}

		if argon2Config, err := config.GenerateArgonSettings(time.Duration(tuneMillis), skipArgon2); err != nil {
			log.Fatalf("Could not generate argon2 settings: %v", err)
		} else if bytes, err := json.Marshal(argon2Config); err != nil {
			log.Fatalf("Coule not marshal argon2 settings: %v", err)
		} else if _, err := configfile.Write(bytes); err != nil {
			log.Fatalf("Could not write to config file %s: %v", path, err)
		} else {
			log.Printf("Successfully wrote to config file to %s", path)
		}
	}
}
