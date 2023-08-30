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

import ApiExplorer from './ApiExplorer';
import { screen, render } from 'src/test-utils';
import { rest } from 'msw';
import { setupServer } from 'msw/node';

const server = setupServer();
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

vi.unmock('react');

describe('ApiExplorer', () => {
    it.skip('displays api documentation', async () => {
        const testApiSpec = {
            openapi: '3.0.3',
            info: {
                title: 'OAS 3.0.3 sample with multiple servers',
                version: '0.1.0',
            },
            servers: [
                {
                    url: 'http://testserver1.com',
                },
                {
                    url: 'http://testserver2.com',
                },
            ],
            paths: {
                '/test/': {
                    get: {
                        responses: {
                            '200': {
                                description: 'Successful Response',
                            },
                        },
                    },
                },
            },
        };

        server.use(
            rest.get('/api/v2/swagger/docs.json', (req, res, ctx) => {
                return res(ctx.json(testApiSpec));
            })
        );

        render(<ApiExplorer />);
        expect(await screen.findByText('API Explorer')).toBeInTheDocument();
    });
});
