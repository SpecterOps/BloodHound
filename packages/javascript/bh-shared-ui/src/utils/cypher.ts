// Copyright 2026 Specter Ops, Inc.
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

/**
 * Wraps a value as a Cypher single-quoted string literal, escaping any
 * embedded backslashes and single quotes so the result is safe to interpolate
 * directly into a Cypher query. The returned value includes the surrounding
 * single quotes.
 *
 * @example
 *   escapeCypherString("a'b")       // "'a\\'b'"
 *   escapeCypherString('S-1-5\\21') // "'S-1-5\\\\21'"
 */
export const escapeCypherString = (value: string): string => `'${value.replace(/\\/g, '\\\\').replace(/'/g, "\\'")}'`;
