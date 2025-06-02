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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from '../../../test-utils';
import ObjectCountPanel from './ObjectCountPanel';

const server = setupServer(
    rest.get(`/api/v2/asset-group-tags/*`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    total_count: 0,
                    counts: [],
                },
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('ObjectCountPanel', () => {
    it('renders error message on error', async () => {
        console.error = vi.fn();
        server.use(
            rest.get('/api/v2/asset-group-tags/1/members/counts', async (_, res, ctx) => {
                return res(ctx.status(403));
            })
        );

        render(<ObjectCountPanel tagId='1' />);

        expect(await screen.findByText('There was an error fetching this data')).toBeInTheDocument();
    });

    it('renders the total count and object counts on success', async () => {
        server.use(
            rest.get('/api/v2/asset-group-tags/1/members/counts', async (_, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            total_count: 100,
                            counts: { 'Object A': 50, 'Object B': 30, 'Object C': 20 },
                        },
                    })
                );
            })
        );

        render(<ObjectCountPanel tagId='1' />);

        expect(screen.getByText('Total Count')).toBeInTheDocument();
        expect(await screen.findByText('100')).toBeInTheDocument();
        expect(screen.getByText('Object A')).toBeInTheDocument();
        expect(screen.getByText('50')).toBeInTheDocument();
        expect(screen.getByText('Object B')).toBeInTheDocument();
        expect(screen.getByText('30')).toBeInTheDocument();
        expect(screen.getByText('Object C')).toBeInTheDocument();
        expect(screen.getByText('20')).toBeInTheDocument();
    });
});
