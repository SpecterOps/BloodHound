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
import { renderHook, waitFor } from '../../test-utils';
import { apiClient } from '../../utils';
import {
    isNodeResponse,
    isRelationshipResponse,
    useGetNodeById,
    useGetRelationshipById,
    useGraphItem,
} from './useGraphItem';

const MOCK_NODE_ID = 42;
const MOCK_CUSTOM_NODE_ID = 43;
const MOCK_USER_NODE_ID = 44;
const MOCK_REL_ID = 99;
const MOCK_OBJECT_ID = 'S-1-5-21-000';

// AZApp is one of the enrichable kinds whose entity endpoint returns a runtime-derived related-node id.
const MOCK_NODE_RESPONSE = {
    node_id: MOCK_NODE_ID,
    kinds: [{ name: 'AZApp', node_kind_id: 1 }],
    properties: { objectid: MOCK_OBJECT_ID },
};

// A node whose kinds are not built in AD/Azure kinds should not be enriched via kind-specific endpoints.
const MOCK_CUSTOM_NODE_RESPONSE = {
    node_id: MOCK_CUSTOM_NODE_ID,
    kinds: [{ name: 'CustomKind', node_kind_id: 2 }],
    properties: { objectid: 'custom-object-id' },
};

// A built in AD kind that is not in the enrichable allow-list should not trigger a kind-specific fetch.
const MOCK_USER_NODE_RESPONSE = {
    node_id: MOCK_USER_NODE_ID,
    kinds: [{ name: 'User', node_kind_id: 4 }],
    properties: { objectid: 'user-object-id' },
};

// Runtime-derived properties that only the kind-specific entity endpoint returns.
const MOCK_ENRICHED_PROPS = { objectid: MOCK_OBJECT_ID, serviceprincipalid: 'SP-123' };

const MOCK_ENRICHED_NODE_RESPONSE = {
    ...MOCK_NODE_RESPONSE,
    properties: { ...MOCK_NODE_RESPONSE.properties, ...MOCK_ENRICHED_PROPS },
};

const MOCK_RELATIONSHIP_RESPONSE = {
    relationship_id: MOCK_REL_ID,
    start_node: 1,
    end_node: 2,
    kind_id: 5,
    properties: {},
};

const server = setupServer(
    rest.get(`/api/v2/nodes/${MOCK_NODE_ID}`, (_req, res, ctx) => {
        return res(ctx.json({ data: MOCK_NODE_RESPONSE }));
    }),
    rest.get(`/api/v2/nodes/${MOCK_CUSTOM_NODE_ID}`, (_req, res, ctx) => {
        return res(ctx.json({ data: MOCK_CUSTOM_NODE_RESPONSE }));
    }),
    rest.get(`/api/v2/nodes/${MOCK_USER_NODE_ID}`, (_req, res, ctx) => {
        return res(ctx.json({ data: MOCK_USER_NODE_RESPONSE }));
    }),
    rest.get('/api/v2/azure/applications', (_req, res, ctx) => {
        return res(ctx.json({ data: { props: MOCK_ENRICHED_PROPS, kinds: ['AZApp', 'AZBase'] } }));
    }),
    rest.get(`/api/v2/relationships/${MOCK_REL_ID}`, (_req, res, ctx) => {
        return res(ctx.json({ data: MOCK_RELATIONSHIP_RESPONSE }));
    })
);

beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
    vi.restoreAllMocks();
});
afterAll(() => server.close());

describe('isRelationshipResponse', () => {
    it('returns true when the response contains relationship_id', () => {
        expect(isRelationshipResponse(MOCK_RELATIONSHIP_RESPONSE as any)).toBe(true);
    });

    it('returns false when the response contains node_id instead', () => {
        expect(isRelationshipResponse(MOCK_NODE_RESPONSE as any)).toBe(false);
    });
});

describe('isNodeResponse', () => {
    it('returns true when the response contains node_id', () => {
        expect(isNodeResponse(MOCK_NODE_RESPONSE as any)).toBe(true);
    });

    it('returns false when the response contains relationship_id instead', () => {
        expect(isNodeResponse(MOCK_RELATIONSHIP_RESPONSE as any)).toBe(false);
    });

    it('returns false when the response is undefined', () => {
        expect(isNodeResponse(undefined)).toBe(false);
    });
});

describe('useGetNodeById', () => {
    it('does not fetch when no id is provided', () => {
        const getNodeByIDSpy = vi.spyOn(apiClient, 'getNodeByID');
        const { result } = renderHook(() => useGetNodeById(undefined));

        expect(getNodeByIDSpy).not.toHaveBeenCalled();
        expect(result.current.isLoading).toBe(false);
        expect(result.current.data).toBeUndefined();
    });

    it('fetches node data when an id is provided', async () => {
        const getNodeByIDSpy = vi.spyOn(apiClient, 'getNodeByID');
        const { result } = renderHook(() => useGetNodeById(MOCK_NODE_ID));

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(getNodeByIDSpy).toHaveBeenCalledWith(
            MOCK_NODE_ID,
            expect.objectContaining({ params: { 'include-info': true } })
        );
        await waitFor(() => expect(result.current.data).toEqual(MOCK_ENRICHED_NODE_RESPONSE));
    });

    it('merges kind-specific runtime properties for enrichable node kinds', async () => {
        const getAZEntityInfoV2Spy = vi.spyOn(apiClient, 'getAZEntityInfoV2');
        const { result } = renderHook(() => useGetNodeById(MOCK_NODE_ID));

        await waitFor(() => expect(result.current.data).toEqual(MOCK_ENRICHED_NODE_RESPONSE));

        expect(getAZEntityInfoV2Spy).toHaveBeenCalledWith(
            'applications',
            MOCK_OBJECT_ID,
            undefined,
            false,
            undefined,
            undefined,
            undefined,
            expect.anything()
        );
        expect(result.current.data?.properties.serviceprincipalid).toBe('SP-123');
    });

    it('falls back to the base node when the kind-specific request fails', async () => {
        server.use(
            rest.get('/api/v2/azure/applications', (_req, res, ctx) => {
                return res(ctx.status(500));
            })
        );

        const { result } = renderHook(() => useGetNodeById(MOCK_NODE_ID));

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(result.current.data).toEqual(MOCK_NODE_RESPONSE);
    });

    it('does not enrich non built in node kinds', async () => {
        const getAZEntityInfoV2Spy = vi.spyOn(apiClient, 'getAZEntityInfoV2');
        const { result } = renderHook(() => useGetNodeById(MOCK_CUSTOM_NODE_ID));

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(getAZEntityInfoV2Spy).not.toHaveBeenCalled();
        expect(result.current.data).toEqual(MOCK_CUSTOM_NODE_RESPONSE);
    });

    it('does not enrich built in node kinds outside the enrichable allow-list', async () => {
        const getAZEntityInfoV2Spy = vi.spyOn(apiClient, 'getAZEntityInfoV2');
        const getUserV2Spy = vi.spyOn(apiClient, 'getUserV2');
        const { result } = renderHook(() => useGetNodeById(MOCK_USER_NODE_ID));

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(getAZEntityInfoV2Spy).not.toHaveBeenCalled();
        expect(getUserV2Spy).not.toHaveBeenCalled();
        expect(result.current.data).toEqual(MOCK_USER_NODE_RESPONSE);
    });
});

describe('useGetRelationshipById', () => {
    it('does not fetch when no id is provided', () => {
        const getRelationshipByIDSpy = vi.spyOn(apiClient, 'getRelationshipByID');
        const { result } = renderHook(() => useGetRelationshipById(undefined));

        expect(getRelationshipByIDSpy).not.toHaveBeenCalled();
        expect(result.current.isLoading).toBe(false);
        expect(result.current.data).toBeUndefined();
    });

    it('fetches relationship data when an id is provided', async () => {
        const getRelationshipByIDSpy = vi.spyOn(apiClient, 'getRelationshipByID');
        const { result } = renderHook(() => useGetRelationshipById(MOCK_REL_ID));

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(getRelationshipByIDSpy).toHaveBeenCalledWith(
            MOCK_REL_ID,
            expect.objectContaining({ params: { 'include-info': true } })
        );
        expect(result.current.data).toEqual(MOCK_RELATIONSHIP_RESPONSE);
    });
});

describe('useGraphItem', () => {
    it('returns undefined query state when itemId is null', () => {
        const { result } = renderHook(() => useGraphItem(null));

        expect(result.current.isLoading).toBe(false);
        expect(result.current.data).toBeUndefined();
    });

    it('returns undefined query state when itemId is undefined', () => {
        const { result } = renderHook(() => useGraphItem(undefined));

        expect(result.current.isLoading).toBe(false);
        expect(result.current.data).toBeUndefined();
    });

    it('fetches node data when given a plain numeric string id', async () => {
        const getNodeByIDSpy = vi.spyOn(apiClient, 'getNodeByID');
        const { result } = renderHook(() => useGraphItem(String(MOCK_NODE_ID)));

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(getNodeByIDSpy).toHaveBeenCalledWith(MOCK_NODE_ID, expect.anything());
        await waitFor(() => expect(result.current.data).toEqual(MOCK_ENRICHED_NODE_RESPONSE));
    });

    it('fetches relationship data when itemId has the rel_ prefix', async () => {
        const getRelationshipByIDSpy = vi.spyOn(apiClient, 'getRelationshipByID');
        const { result } = renderHook(() => useGraphItem(`rel_${MOCK_REL_ID}`));

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(getRelationshipByIDSpy).toHaveBeenCalledWith(MOCK_REL_ID, expect.anything());
        expect(result.current.data).toEqual(MOCK_RELATIONSHIP_RESPONSE);
    });

    it('does not call getRelationshipByID when itemId is a plain node id', () => {
        const getRelationshipByIDSpy = vi.spyOn(apiClient, 'getRelationshipByID');
        renderHook(() => useGraphItem(String(MOCK_NODE_ID)));

        expect(getRelationshipByIDSpy).not.toHaveBeenCalled();
    });

    it('does not call getNodeByID when itemId has the rel_ prefix', () => {
        const getNodeByIDSpy = vi.spyOn(apiClient, 'getNodeByID');
        renderHook(() => useGraphItem(`rel_${MOCK_REL_ID}`));

        expect(getNodeByIDSpy).not.toHaveBeenCalled();
    });
});
