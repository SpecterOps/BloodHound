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

import { formatCompactCount, pluralNodeType } from './graphSimplification';

it('formats natural plurals and compact counts consistently', () => {
    expect(['User', 'AZServicePrincipal', 'Policy', 'Person'].map(pluralNodeType)).toEqual([
        'users',
        'service principals',
        'policies',
        'people',
    ]);
    expect([5, 10100, 300000].map(formatCompactCount)).toEqual(['5', '10.1k', '300k']);
});
