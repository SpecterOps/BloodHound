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
import { useSearchParams } from 'react-router-dom';
import { MultiValueFilterConfig, MultiValueSelection, UseMultiValueFilterParams } from './types';
import { areSelectionsEqual, normalizeSelection } from './utils';

/**
 * Synchronizes a multi-value selection with URL search parameters.
 *
 * - Parameter absence resolves to `defaultSelection`
 * - Repeated `valueParam` values represent `some`
 * - `selectionParam=all` represents explicit `all`
 * - `selectionParam=none` represents explicit `none`
 * - Setting the filter to the `defaultSelection` clears both params
 */
export const useMultiValueFilterParams = ({
    valueParam,
    selectionParam,
    defaultSelection,
}: MultiValueFilterConfig): UseMultiValueFilterParams => {
    const [searchParams, setSearchParams] = useSearchParams();

    // Read from URL params into "selection" on render
    const values = searchParams.getAll(valueParam);
    const currentSelectionKind = searchParams.get(selectionParam);

    const normalizedDefaultSelection = normalizeSelection(defaultSelection);

    let selection = normalizedDefaultSelection;

    if (currentSelectionKind === 'all') {
        selection = { kind: 'all' };
    } else if (currentSelectionKind === 'none') {
        selection = { kind: 'none' };
    } else if (values.length > 0) {
        selection = normalizeSelection({ kind: 'some', values });
    }

    const setSelection = (nextSelection: MultiValueSelection) => {
        const normalizedNextSelection = normalizeSelection(nextSelection);

        setSearchParams((currentParams) => {
            const nextParams = new URLSearchParams(currentParams);

            // Clear both URL parameters before serializing the next selection
            nextParams.delete(valueParam);
            nextParams.delete(selectionParam);

            // If updating to the default value, we should not recreate either param
            if (areSelectionsEqual(normalizedNextSelection, normalizedDefaultSelection)) {
                return nextParams;
            }

            // 'some' creates one param for each value, 'all' or 'none' just create the selection param
            if (normalizedNextSelection.kind === 'some') {
                normalizedNextSelection.values.forEach((value) => {
                    nextParams.append(valueParam, value);
                });
            } else {
                nextParams.set(selectionParam, normalizedNextSelection.kind);
            }

            return nextParams;
        });
    };

    return { selection, setSelection };
};
