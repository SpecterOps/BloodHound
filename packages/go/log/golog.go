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

package log

import golog "log"

// adapter is a simple struct that provides a golang stdlib interface to the BloodHound logging framework.
type adapter struct {
	level Level
}

func (s adapter) Write(msgBytes []byte) (n int, err error) {
	WithLevel(s.level).Msg(string(msgBytes))
	return len(msgBytes), nil
}

// Adapter creates a *golog.Logger instance that will correctly write out structured logs via the BloodHound logging
// framework. This tool is useful when adapting libraries that require the golang stdlib logging interface.
func Adapter(level Level, prefix string, flag int) *golog.Logger {
	return golog.New(adapter{
		level: level,
	}, prefix, flag)
}
