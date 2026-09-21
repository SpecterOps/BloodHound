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

import { act, renderHook } from '../../test-utils';
import { MultiValueFilterConfig } from './types';
import { useMultiValueFilterParams } from './useMultiValueFilterParams';

const TEST_CONFIG = {
    valueParam: 'platform',
    selectionParam: 'platformSelection',
    defaultSelection: { kind: 'all' },
} satisfies MultiValueFilterConfig;

type SetupOptions = {
    route?: string;
    config?: MultiValueFilterConfig;
};

const setup = ({ route = '/', config = TEST_CONFIG }: SetupOptions) => {
    return renderHook(() => useMultiValueFilterParams(config), { route });
};

describe('useMultiValueFilterParams', () => {
    it('returns the configured default selection when the URL contains no filter parameters', () => {
        const { result } = setup({});
        expect(result.current.selection).toEqual({ kind: 'all' });
    });

    it('sorts and dedupes the values in the configured default selection', () => {
        const { result } = setup({
            config: {
                ...TEST_CONFIG,
                defaultSelection: {
                    kind: 'some',
                    values: ['azure', 'active-directory', 'active-directory'],
                },
            },
        });
        expect(result.current.selection).toEqual({ kind: 'some', values: ['active-directory', 'azure'] });
    });

    it('reads an explicit all selection from the URL', () => {
        const { result } = setup({
            route: '/?platformSelection=all',
            config: {
                ...TEST_CONFIG,
                defaultSelection: { kind: 'none' },
            },
        });
        expect(result.current.selection).toEqual({ kind: 'all' });
    });

    it('reads an explicit none selection from the URL', () => {
        const { result } = setup({
            route: '/?platformSelection=none',
            config: {
                ...TEST_CONFIG,
                defaultSelection: { kind: 'all' },
            },
        });
        expect(result.current.selection).toEqual({ kind: 'none' });
    });

    it('reads, sorts, and deduplicates repeated value parameters as a some selection', () => {
        const { result } = setup({ route: '/?platform=azure&platform=active-directory&platform=azure' });
        expect(result.current.selection).toEqual({ kind: 'some', values: ['active-directory', 'azure'] });
    });

    it('reads an explicitly empty value parameter as none', () => {
        const { result } = setup({ route: '/?platform=' });
        expect(result.current.selection).toEqual({ kind: 'none' });
    });

    it('gives an explicit selection parameter precedence over value parameters', () => {
        const { result } = setup({ route: '/?platform=azure&platformSelection=none' });
        expect(result.current.selection).toEqual({ kind: 'none' });
    });

    it('writes a normalized some selection while replacing existing filter parameters and preserving unrelated ones', () => {
        const { result } = setup({
            route: '/?platform=opengraph&platformSelection=none&unrelatedParam=1000',
        });

        act(() => {
            result.current.setSelection({
                kind: 'some',
                values: ['active-directory', 'azure', 'active-directory'],
            });
        });

        const searchParams = new URLSearchParams(window.location.search);

        expect(searchParams.getAll('platform')).toEqual(['active-directory', 'azure']);
        expect(searchParams.has('platformSelection')).toBe(false);
        expect(searchParams.get('unrelatedParam')).toBe('1000');
    });

    it('writes all using the selection parameter and removes existing value parameters', () => {
        const { result } = setup({
            route: '/?platform=windows',
            config: {
                ...TEST_CONFIG,
                defaultSelection: { kind: 'none' },
            },
        });

        act(() => result.current.setSelection({ kind: 'all' }));

        const searchParams = new URLSearchParams(window.location.search);

        expect(searchParams.has('platform')).toBe(false);
        expect(searchParams.get('platformSelection')).toEqual('all');
    });

    it('writes none using the selection parameter and removes existing value parameters', () => {
        const { result } = setup({
            route: '/?platform=windows',
            config: {
                ...TEST_CONFIG,
                defaultSelection: { kind: 'all' },
            },
        });

        act(() => result.current.setSelection({ kind: 'none' }));

        const searchParams = new URLSearchParams(window.location.search);

        expect(searchParams.has('platform')).toBe(false);
        expect(searchParams.get('platformSelection')).toEqual('none');
    });

    it('removes both filter parameters when setting the configured default selection', () => {
        const { result } = setup({
            route: '/?platform=windows&platformSelection=none',
            config: {
                ...TEST_CONFIG,
                defaultSelection: { kind: 'some', values: ['active-directory', 'azure'] },
            },
        });

        act(() =>
            result.current.setSelection({
                kind: 'some',
                values: ['azure', 'active-directory', 'azure'],
            })
        );

        const searchParams = new URLSearchParams(window.location.search);

        expect(searchParams.has('platform')).toBe(false);
        expect(searchParams.has('platformSelection')).toBe(false);
    });
});
