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

package api

import (
	"net/url"
	"strings"
)

func NewJoinedURL(base string, extensions ...string) (string, error) {
	if urlInst, err := url.Parse(base); err != nil {
		return "", err
	} else {
		joinedURL := URLJoinPath(*urlInst, extensions...)
		return joinedURL.String(), nil
	}
}

func URLJoinPath(target url.URL, extensions ...string) url.URL {
	for _, extension := range extensions {
		if strings.HasPrefix(extension, "/") {
			if strings.HasSuffix(target.Path, "/") {
				target.Path = target.Path + extension[1:]
			} else {
				target.Path = target.Path + extension
			}
		} else {
			if strings.HasSuffix(target.Path, "/") {
				target.Path = target.Path + extension
			} else {
				target.Path = target.Path + "/" + extension
			}
		}
	}

	return target
}
