// Copyright 2026 Specter Ops, Inc.
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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from '../../test-utils';
import InfiniteScrollingTable from './InfiniteScrollingTable';

const makeItem = (override: object = {}) => ({
    objectID: 'test-id',
    name: 'Test Node',
    label: 'User',
    ...override,
});

const server = setupServer(
    rest.get('/api/v2/custom-nodes', (req, res, ctx) => {
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

const makeFetchCallback = (items: object[] = [makeItem()]) =>
    vi.fn().mockResolvedValue({ data: items, count: items.length, limit: 128, skip: 0 });

describe('InfiniteScrollingTable', () => {
    it('renders item name after data is fetched', async () => {
        render(<InfiniteScrollingTable fetchDataCallback={makeFetchCallback()} itemCount={1} />);
        expect(await screen.findByText('Test Node')).toBeInTheDocument();
    });

    it('calls onClick with the correct SelectedNode when a row is clicked', async () => {
        const user = userEvent.setup();
        const onClick = vi.fn();

        render(<InfiniteScrollingTable fetchDataCallback={makeFetchCallback()} itemCount={1} onClick={onClick} />);
        await screen.findByText('Test Node');
        await user.click(screen.getByTestId('entity-row'));

        expect(onClick).toHaveBeenCalledWith({ id: 'test-id', name: 'Test Node', type: 'User' });
    });

    it('does not call onClick when no handler is provided', async () => {
        const user = userEvent.setup();

        render(<InfiniteScrollingTable fetchDataCallback={makeFetchCallback()} itemCount={1} />);
        await screen.findByText('Test Node');

        await expect(user.click(screen.getByTestId('entity-row'))).resolves.not.toThrow();
    });

    it('renders rows with pointer cursor when onClick is provided', async () => {
        render(<InfiniteScrollingTable fetchDataCallback={makeFetchCallback()} itemCount={1} onClick={vi.fn()} />);
        await screen.findByText('Test Node');

        expect(screen.getByTestId('entity-row')).toHaveStyle({ cursor: 'pointer' });
    });

    it('renders rows with default cursor when onClick is not provided', async () => {
        render(<InfiniteScrollingTable fetchDataCallback={makeFetchCallback()} itemCount={1} />);
        await screen.findByText('Test Node');

        expect(screen.getByTestId('entity-row')).toHaveStyle({ cursor: 'default' });
    });

    it('normalizes items using props.objectid and props.name as fallbacks', async () => {
        const user = userEvent.setup();
        const onClick = vi.fn();
        const item = { props: { objectid: 'prop-id', name: 'Prop Name' }, kind: 'Computer' };

        render(
            <InfiniteScrollingTable fetchDataCallback={makeFetchCallback([item])} itemCount={1} onClick={onClick} />
        );
        await screen.findByText('Prop Name');
        await user.click(screen.getByTestId('entity-row'));

        expect(onClick).toHaveBeenCalledWith({ id: 'prop-id', name: 'Prop Name', type: 'Computer' });
    });

    it('falls back to objectID as name when name field is absent', async () => {
        const item = { objectID: 'fallback-id', label: 'Group' };

        render(<InfiniteScrollingTable fetchDataCallback={makeFetchCallback([item])} itemCount={1} />);
        expect(await screen.findByText('fallback-id')).toBeInTheDocument();
    });

    it('shows "Unknown" when no name or id fields are available', async () => {
        const item = { label: 'User' };

        render(<InfiniteScrollingTable fetchDataCallback={makeFetchCallback([item])} itemCount={1} />);
        expect(await screen.findByText('Unknown')).toBeInTheDocument();
    });
});
