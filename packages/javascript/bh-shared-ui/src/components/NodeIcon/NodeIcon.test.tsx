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

import { GetCustomNodeKindsResponse } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, render } from '../../test-utils';
import NodeIcon from './NodeIcon';

const icons: GetCustomNodeKindsResponse = {
    data: [
        {
            id: 1,
            kindName: 'KindA',
            config: {
                icon: {
                    type: 'font-awesome',
                    name: 'coffee',
                    color: '#333333',
                },
            },
        },
        {
            id: 1,
            kindName: 'Group',
            config: {
                icon: {
                    type: 'font-awesome',
                    name: 'house',
                    color: '#FFFFFF',
                },
            },
        },
    ],
};

const server = setupServer(
    rest.get(`/api/v2/customnode`, async (_req, res, ctx) => {
        return res(ctx.json(icons));
    })
);
beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
});
afterAll(() => server.close());

describe('NodeIcon', () => {
    const setup = async (nodeType: string) => {
        const screen = await act(async () => {
            return render(<NodeIcon nodeType={nodeType} />);
        });
        return screen;
    };

    it('renders correctly', async () => {
        const screen = await setup('User');
        expect(screen.getByTitle('User')).toBeInTheDocument();
    });

    it('renders correctly when an unexpected nodeType is passed', async () => {
        const testNodeType = 'unexpected value';
        const screen = await setup(testNodeType);
        expect(screen.getByTitle(testNodeType)).toBeInTheDocument();
        expect(screen.getByText('question')).toBeInTheDocument(); // fallback icon
    });

    it('renders custom icon correctly', async () => {
        const screen = await setup('KindA');
        expect(screen.getByTitle('KindA')).toBeInTheDocument();
        expect(screen.getByText('mug-saucer')).toBeInTheDocument();
    });

    it('renders custom icon overlap correctly', async () => {
        const screen = await setup('Group');
        expect(screen.getByTitle('Group')).toBeInTheDocument();
        expect(screen.getByText('house')).toBeInTheDocument();
    });
});
