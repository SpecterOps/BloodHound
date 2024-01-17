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

import { render, screen, waitForElementToBeRemoved } from 'src/test-utils';
import FeatureFlag from './FeatureFlag';
import { rest } from 'msw';
import { setupServer } from 'msw/node';

const server = setupServer(
    rest.get('/api/v2/features', async (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        id: 1,
                        key: 'enabled-flag',
                        name: 'Enabled Flag Name',
                        description: 'Enabled Flag Description',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 2,
                        key: 'disabled-flag',
                        name: 'Disabled Flag Name',
                        description: 'Disabled Flag Description',
                        enabled: false,
                        user_updatable: false,
                    },
                ],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('FeatureFlag', () => {
    it('renders enabled prop when flag is enabled', async () => {
        render(
            <FeatureFlag
                flagKey='enabled-flag'
                disabled={<div>This feature is disabled!</div>}
                enabled={<div>This feature is enabled!</div>}
            />
        );
        expect(await screen.findByText('This feature is enabled!')).toBeInTheDocument();
    });

    it('renders disabled prop when flag is not enabled', async () => {
        render(
            <FeatureFlag
                flagKey='disabled-flag'
                disabled={<div>This feature is disabled!</div>}
                enabled={<div>This feature is enabled!</div>}
            />
        );
        expect(await screen.findByText('This feature is disabled!')).toBeInTheDocument();
    });

    it('renders nothing if flag is enabled and enabled prop is not provided', async () => {
        render(<FeatureFlag flagKey='enabled-flag' />);
        await waitForElementToBeRemoved(() => screen.queryByText('Loading...'));
        expect(document.querySelector('div')?.innerHTML).toEqual('');
    });

    it('renders nothing if flag is disabled and disabled prop is not provided', async () => {
        render(<FeatureFlag flagKey='disabled-flag' />);
        await waitForElementToBeRemoved(() => screen.queryByText('Loading...'));
        expect(document.querySelector('div')?.innerHTML).toEqual('');
    });

    it('renders an error fallback when flag is not found', async () => {
        console.error = vi.fn();
        render(<FeatureFlag flagKey='not-found' errorFallback={<span>Error</span>} />);
        await waitForElementToBeRemoved(() => screen.queryByText('Loading...'));
        expect(screen.getByText('Error')).toBeInTheDocument();
        expect(console.error).toHaveBeenCalledWith('Feature flag "not-found" not found');
    });
});
