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

import { acceptedSearchTypes, parseSearchType } from '../useExploreParams';

describe('useExploreParams', () => {
    describe('parseSearchType', () => {
        it('returns paramValue if it is one of the expected values', () => {
            const expected = acceptedSearchTypes.node;
            const actual = parseSearchType(acceptedSearchTypes.node);
            expect(actual).toEqual(expected);
        });
        it('returns undefined if paramValue is not an expected value', () => {
            const expected = null;
            const actual = parseSearchType('NOT ACCEPTED');
            expect(actual).toEqual(expected);
        });
    });
});
