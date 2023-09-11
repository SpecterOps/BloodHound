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

import { compareTerm } from './OperationsFilterPlugin';

describe('OperationsFilterPlugin', () => {
    it('should return true when tag matches term', () => {
        expect(compareTerm('auth', 'auth')).toEqual(true);
    });

    it('should return true when tag partially matches term', () => {
        expect(compareTerm('auth', 'au')).toEqual(true);
    });

    it('should ignore case when filtering', () => {
        expect(compareTerm('user', 'USER')).toEqual(true);
        expect(compareTerm('USER', 'user')).toEqual(true);
    });

    it('should return false when there is no match', () => {
        expect(compareTerm('user', 'auth')).toEqual(false);
    });
});
