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

package serde

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func ParseURL(rawURL string) (URL, error) {
	if parsedURL, err := url.Parse(rawURL); err != nil {
		return URL{}, fmt.Errorf("raw URL string %s is malformed: %w", rawURL, err)
	} else {
		return FromURLPtr(parsedURL), nil
	}
}

func MustParseURL(rawURL string) URL {
	if parsed, err := ParseURL(rawURL); err != nil {
		panic(fmt.Sprintf("Raw URL string %s is malformed: %v", rawURL, err))
	} else {
		return parsed
	}
}

func FromURL(url url.URL) URL {
	return URL{
		URL: url,
	}
}

func FromURLPtr(url *url.URL) URL {
	return URL{
		URL: *url,
	}
}

type URL struct {
	url.URL
}

func (s URL) AsURL() url.URL {
	return s.URL
}

func (s *URL) unmarshal(buffer []byte, unmarshalFunc func(buffer []byte, target any) error) error {
	var rawURL string

	if err := unmarshalFunc(buffer, &rawURL); err != nil {
		return err
	}

	if parsedURL, err := url.Parse(rawURL); err != nil {
		return err
	} else {
		s.Scheme = parsedURL.Scheme
		s.Opaque = parsedURL.Opaque
		s.User = parsedURL.User
		s.Host = parsedURL.Host
		s.Path = parsedURL.Path
		s.RawPath = parsedURL.RawPath
		s.ForceQuery = parsedURL.ForceQuery
		s.RawQuery = parsedURL.RawQuery
		s.Fragment = parsedURL.Fragment
		s.RawFragment = parsedURL.RawFragment
	}

	return nil
}

func (s *URL) UnmarshalJSON(buffer []byte) error {
	return s.unmarshal(buffer, json.Unmarshal)
}

func (s URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
