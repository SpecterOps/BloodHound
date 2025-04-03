// Copyright 2024 Specter Ops, Inc.
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
import { EntityProperties, FetchEntityPropertiesParams, useFetchEntityProperties } from './useFetchEntityProperties';

const entityObjectIdRequest = () => {
    return rest.get(`/api/v2/${EntityApiPathType}/:id`, async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    props: EntityProperties,
                },
            })
        );
    });
};

const entityGraphIdRequest = () => {
    return rest.post(`/api/v2/graphs/cypher`, async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    edges: [],
                    nodes: [
                        {
                            properties: EntityProperties,
                        },
                    ],
                },
            })
        );
    });
};

const EntityNodeType = 'User' as const;
const EntityApiPathType = 'users' as const;
const EntityGraphId = '5223' as const;
const EntityCustomNodeType = 'Custom' as const;
const EntityProperties: EntityProperties = {
    displayname: 'Steve Draper',
    domain: 'TESTLAB.LOCAL',
    domainsid: 'S-1-5-21-570004220-2248230615-4072641716',
    enabled: true,
    hasspn: true,
    lastlogon: '2019-03-05T16:45:48.253268Z',
    lastlogontimestamp: '2019-03-05T16:45:48.253268Z',
    lastseen: '2024-07-18T17:45:50.805475Z',
    name: 'SteveDraper01962@TESTLAB.LOCAL',
    objectid: 'S-1-5-21-570004220-2248230615-4072641716-5965',
    pwdlastset: '2026-05-17T13:30:00Z',
    system_tags: 'admin_tier_0',
};

describe('useFetchEntityProperties', () => {
    const server = setupServer(entityObjectIdRequest(), entityGraphIdRequest());

    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('Searching for existing node type returns node properties', async () => {
        const initialProps = {
            objectId: EntityProperties.objectid,
            nodeType: EntityNodeType,
        };

        const { result } = renderHook(
            (nodeItemParams: FetchEntityPropertiesParams) => useFetchEntityProperties(nodeItemParams),
            {
                initialProps,
            } as any
        );

        await waitFor(() => {
            expect(result.current.isSuccess).toBe(true);
        });

        expect(result.current.entityProperties).toEqual(EntityProperties);
    });
    it('Searching for custom node type with databaseId returns node properties', async () => {
        const initialProps = {
            objectId: EntityProperties.objectid,
            nodeType: EntityCustomNodeType,
            databaseId: EntityGraphId,
        };

        const { result } = renderHook(
            (nodeItemParams: FetchEntityPropertiesParams) => useFetchEntityProperties(nodeItemParams),
            {
                initialProps,
            } as any
        );

        await waitFor(() => {
            expect(result.current.isSuccess).toBe(true);
        });

        expect(result.current.entityProperties).toEqual(EntityProperties);
    });
});
