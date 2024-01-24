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

package test

import (
	"fmt"
	"github.com/stretchr/testify/require"
)

func RequireNilErr(t Controller, err error) {
	errMsg := ""

	if err != nil {
		errMsg = err.Error()
	}

	require.Nilf(t, err, "Error must be nil but found %T: %s", err, errMsg)
}

func RequireNilErrf(t Controller, err error, format string, parameters ...any) {
	errMsg := ""

	if err != nil {
		errMsg = fmt.Sprintf(format, parameters...)
	}

	require.Nilf(t, err, "Error must be nil but found %T: %s", err, errMsg)
}
