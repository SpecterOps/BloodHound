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
import { AssetGroupTagTypeTier } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { useSearchParams } from 'react-router-dom';
import { renderHook, waitFor } from '../../test-utils';
import { useZoneQueryParams } from './useZoneQueryParams';

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useSearchParams: vi.fn(),
    };
});

const handlers = [
    rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    tags: [
                        { position: 1, id: 42, type: AssetGroupTagTypeTier },
                        { position: 2, id: 23, type: AssetGroupTagTypeTier },
                    ],
                },
            })
        );
    }),
    rest.get('/api/v2/features', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        key: 'tier_management_engine',
                        enabled: true,
                    },
                ],
            })
        );
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('useZoneQueryParams', () => {
    it('returns an object with values for assetGroupTagId, params, and setZoneQueryParams', () => {
        const urlSearchParams = new URLSearchParams();
        urlSearchParams.append('assetGroupTagId', '1');
        vi.mocked(useSearchParams).mockReturnValue([urlSearchParams, vi.fn()]);

        const { result } = renderHook(() => useZoneQueryParams());

        expect(result.current).toHaveProperty('assetGroupTagId');
        expect(result.current).toHaveProperty('params');
        expect(result.current).toHaveProperty('setZoneQueryParams');
    });

    test('the ID value reflects the URL when it is available', () => {
        const urlSearchParams = new URLSearchParams();
        urlSearchParams.append('assetGroupTagId', '777');
        vi.mocked(useSearchParams).mockReturnValue([urlSearchParams, vi.fn()]);

        const { result } = renderHook(() => useZoneQueryParams());

        waitFor(() => {
            expect(result.current.assetGroupTagId).toBe(777);
        });
    });

    test('when there is no ID in the URL the highest position tag ID is returned', async () => {
        const urlSearchParams = new URLSearchParams();
        vi.mocked(useSearchParams).mockReturnValue([urlSearchParams, vi.fn()]);

        const { result } = renderHook(() => useZoneQueryParams());

        await waitFor(() => {
            expect(result.current.assetGroupTagId).toBe(42);
        });
    });

    test('when there is no ID in the URL and no tags available, no ID is returned (it is undefined)', async () => {
        const urlSearchParams = new URLSearchParams();
        vi.mocked(useSearchParams).mockReturnValue([urlSearchParams, vi.fn()]);

        server.use(
            rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            tags: [],
                        },
                    })
                );
            })
        );

        const { result } = renderHook(() => useZoneQueryParams());

        expect(result.current.assetGroupTagId).toBe(undefined);
    });

    it('correctly handles the hygiene ID with value 0 by not omitting it in the params value even though 0 is falsey', async () => {
        const urlSearchParams = new URLSearchParams();
        urlSearchParams.append('assetGroupTagId', '0');
        vi.mocked(useSearchParams).mockReturnValue([urlSearchParams, vi.fn()]);

        const { result } = renderHook(() => useZoneQueryParams());

        await waitFor(() => {
            expect(result.current.assetGroupTagId).toBe(0);
        });
        expect(result.current.params.get('asset_group_tag_id')).toBe('0');
    });
});
