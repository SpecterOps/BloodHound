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

import startCase from 'lodash/startCase';

const compact = new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 1 });
export const formatCompactCount = (count: number): string => compact.format(count).toLowerCase();

export function pluralNodeType(kind: string): string {
    const noun = startCase(kind.replace(/^AZ(?=[A-Z])/, '')).toLowerCase();
    const irregular: Record<string, string> = { person: 'people', child: 'children' };
    if (irregular[noun]) return irregular[noun];
    if (/[^aeiou]y$/.test(noun)) return noun.slice(0, -1) + 'ies';
    return noun + (/(s|x|z|ch|sh)$/.test(noun) ? 'es' : 's');
}
