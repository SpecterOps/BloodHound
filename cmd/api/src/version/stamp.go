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

package version

import (
	"fmt"
	"strconv"
)

var (
	majorVersion      = "999"
	minorVersion      = "999"
	patchVersion      = "999"
	prereleaseVersion = ""
	version           Version
)

// GetVersion returns the current version of the BH application. Since the version is returned as a value instead of a
// reference this should maintain that the unexported version reference remains immutable external to this package.
func GetVersion() Version {
	return version
}

// newVersion returns a newly created Version struct from the stamped variables set by build ldflags. If an error is
// encountered while parsing the strings to integers, the error is returned to the caller alongside an empty Version
// struct.
func newVersion(major, minor, patch, prerelease string) (Version, error) {
	if parsedMajor, err := strconv.Atoi(major); err != nil {
		return Version{}, fmt.Errorf("major version component %s is not a valid integer: %w", major, err)
	} else if parsedMinor, err := strconv.Atoi(minor); err != nil {
		return Version{}, fmt.Errorf("minor version component %s is not a valid integer: %w", minor, err)
	} else if parsedPatch, err := strconv.Atoi(patch); err != nil {
		return Version{}, fmt.Errorf("patch version component %s is not a valid integer: %w", patch, err)
	} else {
		return Version{
			Major:      parsedMajor,
			Minor:      parsedMinor,
			Patch:      parsedPatch,
			Prerelease: prerelease,
		}, nil
	}
}

func init() {
	if newVersion, err := newVersion(majorVersion, minorVersion, patchVersion, prereleaseVersion); err != nil {
		panic(fmt.Sprintf("BloodHound version information incorrect. Please inspect build output and ensure that compiler ldflags have been set correctly: %v", err))
	} else {
		version = newVersion
	}
}
