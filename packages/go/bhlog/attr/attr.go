// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// attr supplies custom slog.Attr constructors
package attr

import "log/slog"

// Error consistently includes an error message via standard logging in the "err" field.
func Error(value error) slog.Attr {
	return slog.String("err", value.Error())
}

// Namespace consistently includes the namespace for a given log via standard logging in the "namespace" field.
// Examples of namespaces include "analysis" for functions executing during the overall analysis process, and "dogtags" for license-related logs.
func Namespace(value string) slog.Attr {
	return slog.String("namespace", value)
}

// Scope consistently includes the scope for a given log via standard logging in the "scope" field.
// Scope was originally created for creating a consistent set of levels of the overall analysis process; the following scopes are used:
// summary: The top-level process that runs all of analysis.
// step: The major steps that run during analysis such as post-processing, tagging, risk analysis and generation, etc.
// process: The processes which run within a given step such as AD post-processing, selecting zone members, etc.
// routine: Any low-level function which runs as part of the process requiring measurement, but not necessarily in the top 3 levels of measurement.
func Scope(value string) slog.Attr {
	return slog.String("scope", value)
}

// Function consistently includes the function name generating the log via standard logging in the "fn" field.
// This should be an exact copy of the Go function name.
func Function(value string) slog.Attr {
	return slog.String("fn", value)
}
