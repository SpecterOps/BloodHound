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

import userEvent from '@testing-library/user-event';
import { NodeDetails } from 'js-client-library';
import { setupServer } from 'msw/node';
import * as hooks from '../../../hooks';
import { mockSourceKindsHandler } from '../../../mocks';
import { render } from '../../../test-utils';
import CopyMenuItem from './CopyMenuItem';

const server = setupServer(mockSourceKindsHandler());

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const useExploreSelectedItemSpy = vi.spyOn(hooks, 'useExploreSelectedItem');

describe('CopyMenuItem', () => {
    const selectedNode: NodeDetails = {
        node_id: 1,
        kinds: [],
        properties: { name: 'foo', objectid: 'bar', lastSeen: '' },
    };

    const setup = () => {
        useExploreSelectedItemSpy.mockReturnValue({ selectedItemQuery: { data: selectedNode } } as any);
        const screen = render(<CopyMenuItem />);
        return screen;
    };

    it('handles copying the name', async () => {
        const screen = setup();

        const user = userEvent.setup();

        const copyOption = screen.getByRole('menuitem', { name: /copy/i });
        await user.click(copyOption);

        const nameOption = await screen.findByRole('menuitem', { name: 'Name' });
        const objectIdOption = screen.getByRole('menuitem', { name: 'Object ID' });
        const cypherOption = screen.getByRole('menuitem', { name: 'Cypher' });

        expect(nameOption).toBeInTheDocument();
        expect(objectIdOption).toBeInTheDocument();
        expect(cypherOption).toBeInTheDocument();

        await user.click(nameOption);

        const clipboardText = await navigator.clipboard.readText();
        expect(clipboardText).toBe(selectedNode.properties.name);
    });
});
