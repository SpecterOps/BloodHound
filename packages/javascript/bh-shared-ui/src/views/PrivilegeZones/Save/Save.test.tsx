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

import { AssetGroupTagTypeOwned, AssetGroupTagTypeZone } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { useParams } from 'react-router-dom';
import Save from '.';
import { detailsPath, privilegeZonesPath, zonesPath } from '../../../routes';
import { render, screen, waitFor } from '../../../test-utils';

const handlers = [
    rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    tags: [
                        { position: 1, id: 42, type: AssetGroupTagTypeZone },
                        { position: 2, id: 23, type: AssetGroupTagTypeZone },
                        { position: 7, id: 1, type: AssetGroupTagTypeZone },
                        { position: 3, id: 2, type: AssetGroupTagTypeZone },
                        { position: 777, id: 3, type: AssetGroupTagTypeZone },
                        { position: null, id: 4, type: AssetGroupTagTypeOwned },
                    ],
                },
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags/1', async (_, res, ctx) => {
        return res(ctx.status(200));
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
    rest.get('api/v2/config', async (req, rest, ctx) => {
        return rest(ctx.status(200));
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useParams: vi.fn(),
    };
});

describe('Create Update pages', () => {
    it('has the correct value for the links in the breadcrumbs', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', labelId: undefined });
        render(<Save />);

        await screen.findByTestId('privilege-zones_save_details-breadcrumb');

        waitFor(async () => {
            expect(screen.getByTestId('privilege-zones_save_details-breadcrumb')).toHaveAttribute(
                'href',
                `${privilegeZonesPath}/${zonesPath}/42/${detailsPath}`
            );
        });
    });
});
