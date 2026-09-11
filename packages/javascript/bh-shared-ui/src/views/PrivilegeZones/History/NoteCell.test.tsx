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
import userEvent from '@testing-library/user-event';
import { BloodHoundString } from 'js-client-library';
import { fireEvent, render, screen } from '../../../test-utils';
import { useHistoryTableContext } from './HistoryTableContext';
import { NoteCell } from './NoteCell';

// Mock AppIcon
vi.mock('../../../components/AppIcon', () => ({
    AppIcon: {
        LinedPaper: (props: any) => <svg data-testid='lined-paper-icon' {...props} />,
    },
}));

// Mock context
vi.mock('./HistoryTableContext', () => ({
    useHistoryTableContext: vi.fn(),
}));

describe('NoteCell component', () => {
    const mockSetSelected = vi.fn();
    const mockClearSelected = vi.fn();

    const defaultItem = {
        id: 132,
        created_at: '2025-10-08T11:00:00.000000Z',
        actor: 'some-user',
        email: 'spam@example.com',
        action: 'CreateSelector',
        target: '6546',
        asset_group_tag_id: 5,
        environment_id: null,
        note: 'note',
        tagName: 'foo',
    };

    beforeEach(() => {
        (useHistoryTableContext as jest.Mock).mockReturnValue({
            selected: null,
            setSelected: mockSetSelected,
            clearSelected: mockClearSelected,
        });
    });

    it('displays the default "No notes" tooltip when the note is missing and actor is not BloodHound', async () => {
        const user = userEvent.setup();

        render(<NoteCell row={{ original: { ...defaultItem, note: null } }} />);

        await user.hover(screen.getByText('-'));

        expect(await screen.findByRole('tooltip', { name: 'No notes' })).toBeInTheDocument();
    });

    it('displays the BloodHound-specific tooltip when the note is missing and actor is BloodHoundString', async () => {
        const user = userEvent.setup();

        render(<NoteCell row={{ original: { ...defaultItem, note: null, actor: BloodHoundString } }} />);

        await user.hover(screen.getByText('-'));

        expect(
            await screen.findByRole('tooltip', { name: `No notes for ${BloodHoundString} history` })
        ).toBeInTheDocument();
    });

    it('renders a button when actor is not BloodHoundString and note exists', () => {
        (useHistoryTableContext as jest.Mock).mockReturnValue({
            selected: defaultItem,
            setSelected: mockSetSelected,
            clearSelected: mockClearSelected,
        });

        render(<NoteCell row={{ original: defaultItem }} />);

        expect(screen.getByRole('button')).toBeInTheDocument();
        expect(screen.getByTestId('lined-paper-icon')).toBeInTheDocument();
    });

    it('calls setSelected with correct data on click', () => {
        (useHistoryTableContext as jest.Mock).mockReturnValue({
            selected: null,
            setSelected: mockSetSelected,
            clearSelected: mockClearSelected,
        });

        render(<NoteCell row={{ original: defaultItem }} />);

        const button = screen.getByRole('button');
        fireEvent.click(button);

        expect(mockSetSelected).toHaveBeenCalledWith(defaultItem);
    });

    it('clears selected if same note is clicked again', () => {
        (useHistoryTableContext as jest.Mock).mockReturnValue({
            selected: defaultItem,
            setSelected: mockSetSelected,
            clearSelected: mockClearSelected,
        });

        render(<NoteCell row={{ original: defaultItem }} />);

        const button = screen.getByRole('button');
        fireEvent.click(button);

        expect(mockClearSelected).toHaveBeenCalled();
    });
});
