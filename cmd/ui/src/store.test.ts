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

describe('redux store', () => {
    beforeEach(() => {
        vi.resetModules();
    });

    test('should clear state when logout action is dispatched', async () => {
        const { rootReducer } = await import('src/store');
        const defaultState = rootReducer(undefined, {});
        // modify the default state
        const previousState: typeof defaultState = {
            ...defaultState,
            assetgroups: {
                ...defaultState.assetgroups,
                assetGroups: ['just', 'some', 'random', 'non-default', 'data'],
            },
        };

        // derive new state
        const newState = rootReducer(previousState, {
            type: 'auth/logout/fulfilled',
        });

        // ensure the new state matches the default state after a logout action is dispatched
        expect(newState).toEqual(defaultState);
    });

    test('app does not crash if "token" cookie exists and "persistedState" localstorage key does not exist', async () => {
        // See https://specterops.atlassian.net/browse/BED-2224 for details
        vi.mock('js-cookie', () => ({
            default: {
                get: () => 'valid_session_token',
                remove: vi.fn(),
            },
        }));
        Storage.prototype.getItem = vi.fn().mockImplementationOnce(() => null);
        await import('src/store');
    });
});

export {};
