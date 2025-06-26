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
import { render, screen, waitFor } from 'src/test-utils';
import GraphView from './GraphView';

const server = setupServer(
    rest.get('/api/v2/graphs/cypher', (req, res, ctx) => {
        return res(ctx.status(500));
    })
);

describe('GraphView', () => {
    it('renders a graph view', () => {
        const { container } = render(<GraphView />);
        expect(container).toBeEmptyDOMElement;
    });

    it('displays an error message', async () => {
        beforeEach(() => server.listen());
        afterAll(() => server.close());
        render(<GraphView />);
        const errorAlert = await waitFor(() =>
            screen.findByText('An unexpected error has occurred. Please refresh the page and try again.')
        );

        expect(errorAlert).toBeInTheDocument();
    });
});
