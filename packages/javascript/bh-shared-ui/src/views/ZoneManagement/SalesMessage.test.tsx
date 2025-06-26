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
import { ConfigurationKey } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from '../../test-utils';
import SalesMessage from './SalesMessage';

const configResponse = {
    data: [
        {
            key: ConfigurationKey.Tiering,
            value: { multi_tier_analysis_enabled: false, tier_limit: 3, label_limit: 10 },
        },
    ],
};

const server = setupServer();

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('SalesMessage', async () => {
    it('renders sales message when multi_tier_analysis_enabled is false', async () => {
        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configResponse));
            })
        );

        render(<SalesMessage />);

        expect(await screen.findByText(/Upgrade Privilege Zones/i)).toBeInTheDocument();
    });

    it('does not render sales message when multi_tier_analysis_enabled is true', () => {
        const configRes = {
            data: [
                {
                    key: ConfigurationKey.Tiering,
                    value: { multi_tier_analysis_enabled: true, tier_limit: 3, label_limit: 10 },
                },
            ],
        };

        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configRes));
            })
        );

        render(<SalesMessage />);

        expect(screen.queryByText(/Upgrade Privilege Zones/i)).not.toBeInTheDocument();
    });
});
