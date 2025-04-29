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

import { queriesAreLoadingOrErrored } from './queries';

describe('queriesAreLoadingOrErrored', () => {
    it.each(['isLoading', 'isError'] as const)('%s returns as expected based on all queries combined', (key) => {
        // all keys to true
        const allTrueActual = queriesAreLoadingOrErrored(
            { [key]: true } as any,
            ...([{ [key]: true }, { [key]: true }] as any),
            ...([{ [key]: true }] as any)
        );
        expect(allTrueActual[key]).toBe(true);

        // Partial keys are true
        const partialTrueActual = queriesAreLoadingOrErrored(
            { [key]: true } as any,
            ...([{ [key]: false }, { [key]: true }] as any),
            ...([{ [key]: false }] as any)
        );
        expect(partialTrueActual[key]).toBe(true);

        // set all keys to false
        const allFalseActual = queriesAreLoadingOrErrored(
            { [key]: false } as any,
            ...([{ [key]: false }, { [key]: false }] as any),
            ...([{ [key]: false }] as any)
        );
        expect(allFalseActual[key]).toBe(false);
    });
});
