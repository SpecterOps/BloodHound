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
import { MultiValueSelection } from './types';

// dedupes and sorts the selected values, also accounts for the ambiguous case of { kind: 'none' } === { kind: 'some', values: [] }
export const normalizeSelection = (selection: MultiValueSelection): MultiValueSelection => {
    if (selection.kind !== 'some') {
        return selection;
    }

    const normalizedValues = [...new Set(selection.values)].filter((value) => value.length > 0).sort();

    if (normalizedValues.length > 0) {
        return { kind: 'some', values: normalizedValues };
    } else {
        return { kind: 'none' };
    }
};

// Strictly compares equality between two selections, arguments should be normalized beforehand
export const areSelectionsEqual = (left: MultiValueSelection, right: MultiValueSelection): boolean => {
    if (left.kind !== right.kind) return false;

    if (left.kind === 'some' && right.kind === 'some') {
        return (
            left.values.length === right.values.length &&
            left.values.every((value, index) => value === right.values[index])
        );
    }

    return true;
};

export const getValuesFromSelection = (selection: MultiValueSelection, options: MultiSelectOption[]) => {
    if (selection.kind === 'some') return selection.values;

    if (selection.kind === 'all') return options.map((option) => option.value);

    if (selection.kind === 'none') return [];

    return [];
};

export const getNextSelection = (newVal: string[], options: MultiSelectOption[]): MultiValueSelection => {
    if (newVal.length === options.length) {
        return { kind: 'all' };
    } else if (newVal.length === 0) {
        return { kind: 'none' };
    } else {
        return {
            kind: 'some',
            values: newVal,
        };
    }
};
