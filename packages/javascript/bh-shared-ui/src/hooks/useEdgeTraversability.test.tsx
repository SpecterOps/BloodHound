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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { renderHook, waitFor } from '../test-utils';
import { useEdgeTraversability } from './useEdgeTraversability';
import { useFeatureFlag } from './useFeatureFlags';

const server = setupServer(
    rest.get('/api/v2/features', (_request, response, context) =>
        response(context.json({ data: [{ key: 'opengraph_extension_management', enabled: true }] }))
    )
);
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('useEdgeTraversability', () => {
    it('uses generated built-in kinds immediately and adds custom kinds from the API', async () => {
        server.use(
            rest.get('/api/v2/extensions-edges', (_request, response, context) =>
                response(
                    context.json({
                        data: [
                            { name: 'AdminTo', is_traversable: true, is_builtin: true },
                            { name: 'Enroll', is_traversable: false, is_builtin: true },
                            { name: 'CustomFact', is_traversable: false, is_builtin: false },
                            { name: 'CustomAttack', is_traversable: true, is_builtin: false },
                        ],
                    })
                )
            )
        );

        const { result } = renderHook(() => useEdgeTraversability());
        expect(result.current.data.get('AdminTo')).toBe(true);
        expect(result.current.data.get('Enroll')).toBe(false);
        expect(result.current.data.get('CustomFact')).toBeUndefined();
        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(result.current.data.get('AdminTo')).toBe(true);
        expect(result.current.data.get('Enroll')).toBe(false);
        expect(result.current.data.get('CustomFact')).toBe(false);
        expect(result.current.data.get('CustomAttack')).toBe(true);
        expect(result.current.data?.get('Unknown')).toBeUndefined();
    });
    it('renders built-in kinds when OpenGraph is disabled without requesting its endpoint', async () => {
        const requestMetadata = vi.fn();
        server.use(
            rest.get('/api/v2/features', (_request, response, context) =>
                response(context.json({ data: [{ key: 'opengraph_extension_management', enabled: false }] }))
            ),
            rest.get('/api/v2/extensions-edges', (_request, response, context) => {
                requestMetadata();
                return response(context.json({ data: [] }));
            })
        );
        const { result } = renderHook(() => ({
            featureQuery: useFeatureFlag('opengraph_extension_management'),
            traversabilityQuery: useEdgeTraversability(),
        }));
        expect(result.current.traversabilityQuery.data.get('Enroll')).toBe(false);
        await waitFor(() => expect(result.current.featureQuery.isSuccess).toBe(true));
        expect(requestMetadata).not.toHaveBeenCalled();
    });
    it('keeps built-in traversability when schema metadata is unavailable', async () => {
        server.use(
            rest.get('/api/v2/extensions-edges', (_request, response, context) => response(context.status(403)))
        );
        const { result } = renderHook(() => useEdgeTraversability());
        await waitFor(() => expect(result.current.isError).toBe(true));
        expect(result.current.data.get('AdminTo')).toBe(true);
        expect(result.current.data.get('Enroll')).toBe(false);
        expect(result.current.data.get('Unknown')).toBeUndefined();
    });
});
