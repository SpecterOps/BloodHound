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

import { createAuthStateWithPermissions } from 'bh-shared-ui';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import {
    ROUTE_ADMINISTRATION_DATA_QUALITY,
    ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES,
    ROUTE_ADMINISTRATION_MANAGE_USERS,
    ROUTE_ADMINISTRATION_SSO_CONFIGURATION,
} from 'src/routes/constants';
import { act, render, screen } from 'src/test-utils';
import Administration from './Administration';

const server = setupServer(
    rest.get('/api/v2/self', (req, res, ctx) => {
        return res(
            ctx.json({
                data: createAuthStateWithPermissions([]).user,
            })
        );
    }),
    rest.get('/api/v2/file-upload/accepted-types', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    'application/json',
                    'application/zip',
                    'application/x-zip-compressed',
                    'application/zip-compressed',
                ],
            })
        );
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        key: 'open_graph_phase_2',
                        enabled: true,
                        user_updatable: true,
                    },
                ],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Administration', () => {
    it('should render navigation links to all Administration sub-pages', async () => {
        await act(async () => render(<Administration />));

        expect(
            screen.getByRole('link', { name: `Navigate to ${ROUTE_ADMINISTRATION_DATA_QUALITY}` })
        ).toBeInTheDocument();
        expect(screen.getByRole('link', { name: `Navigate to ${ROUTE_ADMINISTRATION_DATA_QUALITY}` })).toHaveAttribute(
            'href',
            ROUTE_ADMINISTRATION_DATA_QUALITY
        );

        expect(
            screen.getByRole('link', { name: `Navigate to ${ROUTE_ADMINISTRATION_MANAGE_USERS}` })
        ).toBeInTheDocument();
        expect(screen.getByRole('link', { name: `Navigate to ${ROUTE_ADMINISTRATION_MANAGE_USERS}` })).toHaveAttribute(
            'href',
            ROUTE_ADMINISTRATION_MANAGE_USERS
        );

        expect(
            screen.getByRole('link', { name: `Navigate to ${ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES}` })
        ).toBeInTheDocument();
        expect(
            screen.getByRole('link', { name: `Navigate to ${ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES}` })
        ).toHaveAttribute('href', ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES);

        expect(
            await screen.findByRole('link', { name: `Navigate to ${ROUTE_ADMINISTRATION_SSO_CONFIGURATION}` })
        ).toBeInTheDocument();
        expect(
            screen.getByRole('link', { name: `Navigate to ${ROUTE_ADMINISTRATION_SSO_CONFIGURATION}` })
        ).toHaveAttribute('href', ROUTE_ADMINISTRATION_SSO_CONFIGURATION);
    });

    it('should redirect to nested /file-ingest route if user navigates to /', async () => {
        await act(async () => render(<Administration />));
        expect(window.location.pathname).toBe('/administration/file-ingest');
    });

    it('should redirect to nested /file-ingest route if user navigates to an invalid route', async () => {
        await act(async () =>
            render(<Administration />, {
                route: '/invalid',
            })
        );
        expect(window.location.pathname).toBe('/administration/file-ingest');
    });
});
