// Copyright 2025 Specter Ops, Inc.
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

package fs

import "os"

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/fs.go -package=mocks . Service

// Serves as a lightweight wrapper around the os package which allows for
// path management to be abstracted.
type Service interface {
	CreateTemporaryDirectory(dir, pattern string) (*os.File, error)
}

type Client struct{}

func (c *Client) CreateTemporaryDirectory(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}
