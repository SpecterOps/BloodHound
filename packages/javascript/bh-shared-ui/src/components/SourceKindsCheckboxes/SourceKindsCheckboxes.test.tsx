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
import { SourceKindsCheckboxes } from './SourceKindsCheckboxes';

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

describe('SourceKindsCheckboxes', () => {
    const server = setupServer(
        rest.get('/api/v2/graphs/source-kinds', (req, res, ctx) => {
            return res(ctx.json(SOURCE_KINDS_RESPONSE));
        })
    );

    const defaultProps = {
        checked: [],
        onChange: vi.fn(),
    };

    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('displays loading state', async () => {
        render(<SourceKindsCheckboxes {...defaultProps} />);
        expect(screen.getAllByRole('status')).toHaveLength(3);
    });

    it('hides loading state when data is available', async () => {
        await act(async () => render(<SourceKindsCheckboxes {...defaultProps} />));
        expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });

    it('disables checkboxes by default', async () => {
        await act(async () => render(<SourceKindsCheckboxes {...defaultProps} />));
        const disabled = screen.getAllByRole('checkbox').filter((item) => (item as HTMLInputElement).disabled);

        // Top level and 4 mock children checkboxes after loading
        expect(disabled).toHaveLength(5);
    });

    it('disables checkboxes while loading', async () => {
        render(<SourceKindsCheckboxes {...defaultProps} disabled={false} />);
        const disabled = screen.getAllByRole('checkbox').filter((item) => (item as HTMLInputElement).disabled);

        // "All graph data" and 3 loading state checkboxes
        expect(disabled).toHaveLength(4);
    });

    it('has no checked boxes by default', async () => {
        await act(async () => render(<SourceKindsCheckboxes {...defaultProps} />));
        const checked = screen.queryByRole('checkbox', { checked: true });
        expect(checked).not.toBeInTheDocument();
    });

    it('shows some checked boxes', async () => {
        await act(async () => render(<SourceKindsCheckboxes {...defaultProps} checked={[1, 2]} />));

        const checked = screen.getAllByRole('checkbox', { checked: true });
        expect(checked).toHaveLength(2);

        const parent = screen.getAllByRole('checkbox', { checked: false })[0];
        expect(parent).toHaveAttribute('data-indeterminate', 'true');
    });

    it('shows all checked boxes', async () => {
        await act(async () => render(<SourceKindsCheckboxes {...defaultProps} checked={[0, 1, 2, 3]} />));
        const checked = screen.getAllByRole('checkbox', { checked: true });

        // Parent and 4 checked children
        expect(checked).toHaveLength(5);
    });

    it('toggles parent from all to none', async () => {
        const onChange = vi.fn();

        await act(async () =>
            render(
                <SourceKindsCheckboxes {...defaultProps} disabled={false} checked={[0, 1, 2, 3]} onChange={onChange} />
            )
        );

        const user = userEvent.setup();
        const parent = screen.getAllByRole('checkbox')[0];
        await user.click(parent);

        // Component `checked` state update is controlled from parent
        // Instead of testing fo new checked state, test onChange args
        expect(onChange).toHaveBeenCalledWith([]);
    });

    it('toggles parent from none to all', async () => {
        const onChange = vi.fn();

        await act(async () =>
            render(<SourceKindsCheckboxes {...defaultProps} disabled={false} checked={[]} onChange={onChange} />)
        );

        const user = userEvent.setup();
        const parent = screen.getAllByRole('checkbox')[0];
        await user.click(parent);

        // Component `checked` state update is controlled from parent
        // Instead of testing fo new checked state, test onChange args
        expect(onChange).toHaveBeenCalledWith([1, 2, 3, 0]);
    });

    it('toggles parent from some to all', async () => {
        const onChange = vi.fn();

        await act(async () =>
            render(<SourceKindsCheckboxes {...defaultProps} disabled={false} checked={[1, 2]} onChange={onChange} />)
        );

        const user = userEvent.setup();
        const parent = screen.getAllByRole('checkbox')[0];
        await user.click(parent);

        // Component `checked` state update is controlled from parent
        // Instead of testing fo new checked state, test onChange args
        expect(onChange).toHaveBeenCalledWith([1, 2, 3, 0]);
    });

    it('toggles a child on', async () => {
        const onChange = vi.fn();

        await act(async () => render(<SourceKindsCheckboxes {...defaultProps} disabled={false} onChange={onChange} />));

        const user = userEvent.setup();
        const parent = screen.getAllByRole('checkbox')[1];
        await user.click(parent);

        // Component `checked` state update is controlled from parent
        // Instead of testing fo new checked state, test onChange args
        expect(onChange).toHaveBeenCalledWith([1]);
    });

    it('toggles a child off', async () => {
        const onChange = vi.fn();

        await act(async () =>
            render(<SourceKindsCheckboxes {...defaultProps} checked={[1, 2]} disabled={false} onChange={onChange} />)
        );

        const user = userEvent.setup();
        const parent = screen.getAllByRole('checkbox')[1];
        await user.click(parent);

        // Component `checked` state update is controlled from parent
        // Instead of testing fo new checked state, test onChange args
        expect(onChange).toHaveBeenCalledWith([2]);
    });
});
