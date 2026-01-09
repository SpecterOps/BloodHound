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

// One or more non-whitespace characters on both sides of a period
const domainRegex = /\S+\.\S+/u;

/**
 * Returns true if the string looks like it might be a domain,
 * i.e. non-whitespace characters on both sides of a period,
 *
 * @param str - string to test
 *
 * @returns true if the string looks like a domain
 */
export function isDomainLike(str: string): boolean {
    return domainRegex.test(str);
}
