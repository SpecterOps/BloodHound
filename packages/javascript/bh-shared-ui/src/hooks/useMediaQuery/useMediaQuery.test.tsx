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
import { createMatchMediaController } from '../../testing';
import { useMediaQuery } from './useMediaQuery';

const query = '(min-width: 1280px)';
const alternateQuery = '(prefers-color-scheme: dark)';

describe('useMediaQuery', () => {
    afterEach(() => vi.unstubAllGlobals());

    it('returns the current media-query match on initial render', () => {
        createMatchMediaController(true);

        const { result } = renderHook(() => useMediaQuery(query));

        expect(result.current).toBe(true);
    });

    it('updates when the media-query match changes', () => {
        const matchMediaController = createMatchMediaController(false);
        const { result } = renderHook(() => useMediaQuery(query));

        act(() => matchMediaController.setMatches(true));

        expect(result.current).toBe(true);
    });

    it('keeps multiple media queries independent', () => {
        const matchMediaController = createMatchMediaController({
            [query]: false,
            [alternateQuery]: true,
        });
        const { result } = renderHook(() => [useMediaQuery(query), useMediaQuery(alternateQuery)]);

        act(() => matchMediaController.setMatches(true, query));

        expect(result.current).toEqual([true, true]);
    });

    it('removes its change listener when unmounted', () => {
        const matchMediaController = createMatchMediaController(false);
        const { unmount } = renderHook(() => useMediaQuery(query));

        unmount();

        expect(matchMediaController.getMediaQueryList(query).removeEventListener).toHaveBeenCalledWith(
            'change',
            expect.any(Function)
        );
    });
});
