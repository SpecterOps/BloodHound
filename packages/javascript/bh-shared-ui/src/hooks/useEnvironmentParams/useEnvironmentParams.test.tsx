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

import { environmentAggregationMap, parseEnvironmentAggregation } from './useEnvironmentParams';

describe('useEnvironmentParams', () => {
    describe('parseEnvironmentAggregation', () => {
        it.each([0, null, undefined, ''])('returns null if paramValue is %o', (paramValue) => {
            const expected = null;
            const actual = parseEnvironmentAggregation(paramValue as any);
            expect(actual).toBe(expected);
        });
        it.each(Object.keys(environmentAggregationMap))(
            'returns paramValue as EnvironmentAggregation if paramValue is %s',
            (paramValue) => {
                const expected = paramValue;
                const actual = parseEnvironmentAggregation(paramValue as any);
                expect(actual).toBe(expected);
            }
        );
        it('returns null if paramValue is not an EnvironmentAggregation', () => {
            const expected = null;
            const actual = parseEnvironmentAggregation('DEFINITELY_SHOULDNT_BE_PARAM_EVER' as any);
            expect(actual).toBe(expected);
        });
    });
});
