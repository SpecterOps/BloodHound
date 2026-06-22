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

import { setupServer } from 'msw/node';
import { ActiveDirectoryNodeKind } from '../../graphSchema';
import { mockKindsHandler } from '../../mocks/handlers';
import { renderHook, waitFor } from '../../test-utils';
import { getEmptyResultsText, useKeywordAndTypeValues } from './useSearch';

const server = setupServer(mockKindsHandler(['computer', 'azapp']));

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Getting the text for the disabled item display for a search when there are no results', () => {
    describe('Loading states', () => {
        test('The data is loading', () => {
            expect(getEmptyResultsText(true, false, false, {}, 'test', undefined, 'test', [])).toEqual('Loading...');
        });

        test('The data is fetching', () => {
            expect(getEmptyResultsText(false, true, false, {}, 'test', undefined, 'test', [])).toEqual('Loading...');
        });
    });

    describe('Error states', () => {
        test('Fetching the data timed out', () => {
            expect(
                getEmptyResultsText(false, false, true, { response: { status: 504 } }, 'test', undefined, 'test', [])
            ).toEqual('Search has timed out. Please try again.');
        });

        test('An unknown error occurred when fetching the data', () => {
            expect(
                getEmptyResultsText(false, false, true, { details: 'mock error' }, 'test', undefined, 'test', [])
            ).toEqual('An error has occurred. Please try again.');
        });
    });

    describe('Empty data states', () => {
        test('The user has yet to type into the search', () => {
            expect(getEmptyResultsText(false, false, false, {}, '', undefined, '', [])).toEqual(
                'Begin typing to search. Prepend a type followed by a colon to search by type, e.g., user:bob'
            );
        });

        test('The user has yet to type into the search (for OU and Domain searches that have hard coded types)', () => {
            expect(getEmptyResultsText(false, false, false, {}, '', ActiveDirectoryNodeKind.OU, '', [])).toEqual(
                'Begin typing to search.'
            );
        });

        test('A type was provided but no keyword follows the typed filter', () => {
            expect(
                getEmptyResultsText(false, false, false, {}, 'computer:', ActiveDirectoryNodeKind.Computer, '', [])
            ).toEqual('Include a keyword to search for a node of type Computer');
        });

        test('A type and keyword were provided but there were no results available', () => {
            expect(
                getEmptyResultsText(
                    false,
                    false,
                    false,
                    {},
                    'container:test',
                    ActiveDirectoryNodeKind.Container,
                    'test',
                    []
                )
            ).toEqual('No results found for "test" of type Container');
        });

        test('No type was provided and no results were available', () => {
            expect(getEmptyResultsText(false, false, false, {}, 'test', undefined, 'test', [])).toEqual(
                'No results found for "test"'
            );
        });
    });
});

describe('Parsing the debounced input for type and keyword values', () => {
    test('`undefined` input is provided', () => {
        const { result } = renderHook(() => useKeywordAndTypeValues(undefined));
        expect(result.current).toEqual({ keyword: undefined, type: undefined });
    });

    test('Empty input is provided', () => {
        const { result } = renderHook(() => useKeywordAndTypeValues(''));
        expect(result.current).toEqual({ keyword: '', type: undefined });
    });

    test('Input does not contain a type', () => {
        const { result } = renderHook(() => useKeywordAndTypeValues('test'));
        expect(result.current).toEqual({ keyword: 'test', type: undefined });
    });

    it('Will treat what is to the left of the first colon as a search filter if it matches an existing kind', async () => {
        const { result } = renderHook(() => useKeywordAndTypeValues('computer:user:domain:ou:gpo:test'));
        await waitFor(() => expect(result.current.type).toBe('computer'));
        expect(result.current).toEqual({ keyword: 'user:domain:ou:gpo:test', type: 'computer' });
    });

    it('Will ignore what is to the left of the first colon if it doesnt match an existing kind and will treat all as keyword', async () => {
        const { result } = renderHook(() => useKeywordAndTypeValues('testing:user:domain:ou:gpo:test'));
        await waitFor(() => expect(result.current.keyword).toBe('testing:user:domain:ou:gpo:test'));
        expect(result.current).toEqual({ keyword: 'testing:user:domain:ou:gpo:test', type: undefined });
    });
});
