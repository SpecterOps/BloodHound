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

import { type MultiSelectOption } from 'doodle-ui';
import { getNextSelection, getValuesFromSelection } from './utils';

const TEST_OPTIONS = [
    { value: 'apl', label: 'Apple' },
    { value: 'bnana', label: 'Banana' },
    { value: 'ccmbr', label: 'Cucumber' },
] satisfies MultiSelectOption[];

describe('getValuesFromSelection', () => {
    it('returns the selected values for a some selection', () => {
        expect(getValuesFromSelection({ kind: 'some', values: ['apl', 'bnana'] }, TEST_OPTIONS)).toEqual([
            'apl',
            'bnana',
        ]);
    });

    it('returns the values of every option for an all selection', () => {
        expect(getValuesFromSelection({ kind: 'all' }, TEST_OPTIONS)).toEqual(['apl', 'bnana', 'ccmbr']);
    });

    it('returns an empty array for an all selection when there are no options', () => {
        expect(getValuesFromSelection({ kind: 'all' }, [])).toEqual([]);
    });

    it('returns an empty array for a none selection', () => {
        expect(getValuesFromSelection({ kind: 'none' }, TEST_OPTIONS)).toEqual([]);
    });

    it('returns an empty array for a some selection with no values', () => {
        expect(getValuesFromSelection({ kind: 'some', values: [] }, TEST_OPTIONS)).toEqual([]);
    });
});

describe('getNextSelection', () => {
    it('returns an all selection when every option is selected', () => {
        expect(getNextSelection(['apl', 'bnana', 'gccmbr'], TEST_OPTIONS)).toEqual({ kind: 'all' });
    });

    it('returns a none selection when no options are selected', () => {
        expect(getNextSelection([], TEST_OPTIONS)).toEqual({ kind: 'none' });
    });

    it('returns a some selection with the selected values when a subset of options is selected', () => {
        expect(getNextSelection(['bnana'], TEST_OPTIONS)).toEqual({ kind: 'some', values: ['bnana'] });
    });

    it('returns an all selection when both the new value and the options are empty', () => {
        expect(getNextSelection([], [])).toEqual({ kind: 'all' });
    });
});
