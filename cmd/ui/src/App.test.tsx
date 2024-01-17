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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import App from 'src/App';
import { render, screen } from 'src/test-utils';

const server = setupServer(
    rest.get('/api/v2/saml/sso', (req, res, ctx) => {
        return res(
            ctx.json({
                endpoints: [],
            })
        );
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// override manual mock
vi.unmock('react');

describe('app', () => {
    it.skip('renders', async () => {
        render(<App />);
        expect(await screen.findByText('LOGIN')).toBeInTheDocument();
    });
});
