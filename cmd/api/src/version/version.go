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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/specterops/bloodhound/errors"
)

var (
	// See: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	semverParsingRegex = regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
)

const (
	Prefix = "v"

	// Capture group indices for the semver parsing regex
	majorCaptureGroup      = 1
	minorCaptureGroup      = 2
	patchCaptureGroup      = 3
	prereleaseCaptureGroup = 4
)

type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
}

func (s *Version) IsPrerelease() bool {
	return s.Prerelease != ""
}

func (s *Version) unmarshal(buffer []byte, unmarshalFunc func(buffer []byte, target any) error) error {
	var rawValue string

	if err := unmarshalFunc(buffer, &rawValue); err != nil {
		return err
	}

	if parsedVersion, err := Parse(rawValue); err != nil {
		return err
	} else {
		s.Major = parsedVersion.Major
		s.Minor = parsedVersion.Minor
		s.Patch = parsedVersion.Patch
		s.Prerelease = parsedVersion.Prerelease
	}

	return nil
}

func (s *Version) UnmarshalJSON(buffer []byte) error {
	return s.unmarshal(buffer, json.Unmarshal)
}

func (s Version) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s Version) LessThan(other Version) bool {
	return s.Major < other.Major ||
		(s.Major == other.Major && (s.Minor < other.Minor ||
			(s.Minor == other.Minor && s.Patch < other.Patch)))
}

func (s Version) GreaterThan(other Version) bool {
	return s.Major > other.Major ||
		(s.Major == other.Major && (s.Minor > other.Minor ||
			(s.Minor == other.Minor && s.Patch > other.Patch)))
}

func (s Version) Equals(other Version) bool {
	return s.Major == other.Major && s.Minor == other.Minor && s.Patch == other.Patch
}

func (s Version) String() string {
	if s.Prerelease != "" {
		return fmt.Sprintf("%s%d.%d.%d-%s", Prefix, s.Major, s.Minor, s.Patch, s.Prerelease)
	}

	return fmt.Sprintf("%s%d.%d.%d", Prefix, s.Major, s.Minor, s.Patch)
}

func Parse(rawVersion string) (Version, error) {
	if !strings.HasPrefix(rawVersion, Prefix) {
		return Version{}, fmt.Errorf("version string %s does not start with the prefix %s", rawVersion, Prefix)
	}

	if matches := semverParsingRegex.FindAllStringSubmatch(rawVersion[1:], 1); len(matches) != 1 {
		return Version{}, errors.Error("expected version to be formatted: <major>.<minor>.<patch>[-<prerelease>]")
	} else {
		// Map to the first set of capture groups
		versionParts := matches[0]

		return newVersion(versionParts[majorCaptureGroup], versionParts[minorCaptureGroup], versionParts[patchCaptureGroup], versionParts[prereleaseCaptureGroup])
	}
}
