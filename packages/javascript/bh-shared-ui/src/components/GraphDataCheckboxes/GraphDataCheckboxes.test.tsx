// Copyright 2024 Specter Ops, Inc.
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
import { act, render, screen } from '../../test-utils';
import { GraphDataCheckboxes, type GraphDataSelections } from './GraphDataCheckboxes';

const SOURCE_KINDS_RESPONSE = {
    data: {
        kinds: [
            {
                id: 1,
                name: 'Base',
            },
            {
                id: 2,
                name: 'AZBase',
            },
            {
                id: 3,
                name: 'ACustomBase',
            },
            {
                id: 0,
                name: 'Sourceless',
            },
        ],
    },
};

const ACTIVE_DIRECTORY_DATA_ONLY: GraphDataSelections = {
    sourceKinds: [1],
    relationships: [],
};

const ACTIVE_DIRECTORY_ALL_CHECKED: GraphDataSelections = {
    sourceKinds: [1],
    relationships: ['HasSession'],
};

const ALL_GRAPH_DATA_CHECKED: GraphDataSelections = {
    sourceKinds: [0, 1, 2, 3],
    relationships: ['HasSession'],
};

const HAS_SESSION_ONLY: GraphDataSelections = {
    sourceKinds: [],
    relationships: ['HasSession'],
};

describe('GraphDataCheckboxes', () => {
    const server = setupServer(
        rest.get('/api/v2/graphs/source-kinds', (_req, res, ctx) => {
            return res(ctx.json(SOURCE_KINDS_RESPONSE));
        })
    );

    const defaultProps = {
        checkedSourceKinds: [],
        checkedRelationships: [],
        onChange: vi.fn(),
    };

    beforeAll(() => server.listen());
    afterEach(() => {
        server.resetHandlers();
        vi.clearAllMocks();
    });
    afterAll(() => server.close());

    it('displays loading state', async () => {
        render(<GraphDataCheckboxes {...defaultProps} />);
        expect(screen.getAllByRole('status')).toHaveLength(3);
    });

    it('hides loading state when data is available', async () => {
        await act(async () => render(<GraphDataCheckboxes {...defaultProps} />));
        expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });

    it('disables checkboxes by default', async () => {
        await act(async () => render(<GraphDataCheckboxes {...defaultProps} />));
        const disabled = screen.getAllByRole('checkbox').filter((item) => (item as HTMLInputElement).disabled);

        // Top level, 4 source kind checkboxes, and the HasSession child checkbox
        expect(disabled).toHaveLength(6);
    });

    it('disables checkboxes while loading', async () => {
        render(<GraphDataCheckboxes {...defaultProps} disabled={false} />);
        const disabled = screen.getAllByRole('checkbox').filter((item) => (item as HTMLInputElement).disabled);

        // "All graph data" and 3 loading state checkboxes
        expect(disabled).toHaveLength(4);
    });

    it('has no checked boxes by default', async () => {
        await act(async () => render(<GraphDataCheckboxes {...defaultProps} />));
        const checked = screen.queryByRole('checkbox', { checked: true });
        expect(checked).not.toBeInTheDocument();
    });

    it('shows indeterminate states when some graph data is selected', async () => {
        await act(async () =>
            render(
                <GraphDataCheckboxes
                    {...defaultProps}
                    checkedSourceKinds={ACTIVE_DIRECTORY_DATA_ONLY.sourceKinds}
                    checkedRelationships={ACTIVE_DIRECTORY_DATA_ONLY.relationships}
                />
            )
        );

        expect(screen.getByRole('checkbox', { name: 'All graph data' })).toHaveAttribute('data-indeterminate', 'true');
        expect(screen.getByRole('checkbox', { name: 'Active Directory data' })).toBeChecked();
        expect(screen.getByRole('checkbox', { name: /HasSession/i })).toBeChecked();
    });

    it('shows all checked boxes', async () => {
        await act(async () =>
            render(
                <GraphDataCheckboxes
                    {...defaultProps}
                    checkedSourceKinds={ALL_GRAPH_DATA_CHECKED.sourceKinds}
                    checkedRelationships={ALL_GRAPH_DATA_CHECKED.relationships}
                />
            )
        );
        const checked = screen.getAllByRole('checkbox', { checked: true });

        // Parent, 4 source kind checkboxes, and the HasSession child checkbox
        expect(checked).toHaveLength(6);
    });

    it('toggles parent from all to none', async () => {
        const onChange = vi.fn();

        await act(async () =>
            render(
                <GraphDataCheckboxes
                    {...defaultProps}
                    disabled={false}
                    checkedSourceKinds={ALL_GRAPH_DATA_CHECKED.sourceKinds}
                    checkedRelationships={ALL_GRAPH_DATA_CHECKED.relationships}
                    onChange={onChange}
                />
            )
        );

        const user = userEvent.setup();
        const parent = screen.getByRole('checkbox', { name: 'All graph data' });
        await user.click(parent);

        // Component `checked` state update is controlled from parent
        // Instead of testing for new checked state, test onChange args
        expect(onChange).toHaveBeenCalledWith({ sourceKinds: [], relationships: [] });
    });

    it('toggles parent from none to all', async () => {
        const onChange = vi.fn();

        await act(async () =>
            render(
                <GraphDataCheckboxes
                    {...defaultProps}
                    disabled={false}
                    checkedSourceKinds={[]}
                    checkedRelationships={[]}
                    onChange={onChange}
                />
            )
        );

        const user = userEvent.setup();
        const parent = screen.getByRole('checkbox', { name: 'All graph data' });
        await user.click(parent);

        // Component `checked` state update is controlled from parent
        expect(onChange).toHaveBeenCalledWith(ALL_GRAPH_DATA_CHECKED);
    });

    it('toggles parent from some to all', async () => {
        const onChange = vi.fn();

        await act(async () =>
            render(
                <GraphDataCheckboxes
                    {...defaultProps}
                    disabled={false}
                    checkedSourceKinds={ACTIVE_DIRECTORY_DATA_ONLY.sourceKinds}
                    checkedRelationships={ACTIVE_DIRECTORY_DATA_ONLY.relationships}
                    onChange={onChange}
                />
            )
        );

        const user = userEvent.setup();
        const parent = screen.getByRole('checkbox', { name: 'All graph data' });
        await user.click(parent);

        // Component `checked` state update is controlled from parent
        expect(onChange).toHaveBeenCalledWith(ALL_GRAPH_DATA_CHECKED);
    });

    it('toggles a source kind on with all nested graph data options', async () => {
        const onChange = vi.fn();

        await act(async () => render(<GraphDataCheckboxes {...defaultProps} disabled={false} onChange={onChange} />));

        const user = userEvent.setup();
        const parent = screen.getByRole('checkbox', { name: 'Active Directory data' });
        await user.click(parent);

        // Component `checked` state update is controlled from parent
        expect(onChange).toHaveBeenCalledWith(ACTIVE_DIRECTORY_ALL_CHECKED);
    });

    it('toggles a source kind off from all selected nested graph data options', async () => {
        const onChange = vi.fn();

        await act(async () =>
            render(
                <GraphDataCheckboxes
                    {...defaultProps}
                    checkedSourceKinds={ACTIVE_DIRECTORY_ALL_CHECKED.sourceKinds}
                    checkedRelationships={ACTIVE_DIRECTORY_ALL_CHECKED.relationships}
                    disabled={false}
                    onChange={onChange}
                />
            )
        );

        const user = userEvent.setup();
        const parent = screen.getByRole('checkbox', { name: 'Active Directory data' });
        await user.click(parent);

        // Component `checked` state update is controlled from parent
        expect(onChange).toHaveBeenCalledWith({ sourceKinds: [], relationships: [] });
    });

    it('toggles a nested graph data option off', async () => {
        const onChange = vi.fn();

        await act(async () =>
            render(
                <GraphDataCheckboxes
                    {...defaultProps}
                    checkedSourceKinds={HAS_SESSION_ONLY.sourceKinds}
                    checkedRelationships={HAS_SESSION_ONLY.relationships}
                    disabled={false}
                    onChange={onChange}
                />
            )
        );

        const user = userEvent.setup();
        const child = screen.getByRole('checkbox', { name: /HasSession/i });
        await user.click(child);

        expect(onChange).toHaveBeenCalledWith({ sourceKinds: [], relationships: [] });
    });

    it('does not change a nested graph data option while its source kind is selected', async () => {
        const onChange = vi.fn();

        await act(async () =>
            render(
                <GraphDataCheckboxes
                    {...defaultProps}
                    checkedSourceKinds={ACTIVE_DIRECTORY_DATA_ONLY.sourceKinds}
                    checkedRelationships={ACTIVE_DIRECTORY_DATA_ONLY.relationships}
                    disabled={false}
                    onChange={onChange}
                />
            )
        );

        const user = userEvent.setup();
        const child = screen.getByRole('checkbox', { name: /HasSession/i });
        await user.click(child);

        expect(onChange).not.toHaveBeenCalled();
    });
});
