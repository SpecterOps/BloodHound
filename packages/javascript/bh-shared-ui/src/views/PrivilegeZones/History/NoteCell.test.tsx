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
import { SystemString } from 'js-client-library';
import { fireEvent, render, screen } from '../../../test-utils';
import { HistoryNote, useHistoryTableContext } from './HistoryTableContext';
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
    const mockSetCurrentNote = vi.fn();
    const mockClearCurrentNote = vi.fn();

    const defaultNoteData = {
        email: 'user@example.com',
        note: 'This is a note',
        date: '2025-10-08 11:00:00',
        actor: 'some-user',
    };

    it('renders a dash when actor is SystemString', () => {
        (useHistoryTableContext as jest.Mock).mockReturnValue({
            currentNote: null,
            setCurrentNote: mockSetCurrentNote,
            isCurrentNote: vi.fn(),
            clearCurrentNote: mockClearCurrentNote,
        });

        render(<NoteCell row={{ original: { ...defaultNoteData, actor: SystemString } }} />);

        expect(screen.getByText('-')).toBeInTheDocument();
        expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('renders a button when actor is not SystemString and note exists', () => {
        const currentNote: HistoryNote = {
            note: defaultNoteData.note,
            createdBy: defaultNoteData.email,
            timestamp: defaultNoteData.date,
        };

        (useHistoryTableContext as jest.Mock).mockReturnValue({
            currentNote,
            setCurrentNote: mockSetCurrentNote,
            isCurrentNote: vi.fn(),
            clearCurrentNote: mockClearCurrentNote,
        });

        render(<NoteCell row={{ original: defaultNoteData }} />);

        expect(screen.getByRole('button')).toBeInTheDocument();
        expect(screen.getByTestId('lined-paper-icon')).toBeInTheDocument();
    });

    it('button is disabled when note is falsy', () => {
        (useHistoryTableContext as jest.Mock).mockReturnValue({
            currentNote: null,
            setCurrentNote: mockSetCurrentNote,
            isCurrentNote: () => false,
            clearCurrentNote: mockClearCurrentNote,
        });

        const rowData = { ...defaultNoteData, note: null };

        render(<NoteCell row={{ original: rowData }} />);

        const button = screen.getByRole('button');
        expect(button).toBeDisabled();
    });

    it('calls setCurrentNote with correct data on click', () => {
        (useHistoryTableContext as jest.Mock).mockReturnValue({
            currentNote: null,
            setCurrentNote: mockSetCurrentNote,
            isCurrentNote: () => false,
            clearCurrentNote: mockClearCurrentNote,
        });

        render(<NoteCell row={{ original: defaultNoteData }} />);

        const button = screen.getByRole('button');
        fireEvent.click(button);

        expect(mockSetCurrentNote).toHaveBeenCalledWith({
            note: defaultNoteData.note,
            createdBy: defaultNoteData.email,
            timestamp: defaultNoteData.date,
        });
    });

    it('clears currentNote if same note is clicked again', () => {
        const currentNote = {
            note: defaultNoteData.note,
            createdBy: defaultNoteData.email,
            timestamp: defaultNoteData.date,
        };

        (useHistoryTableContext as jest.Mock).mockReturnValue({
            currentNote,
            setCurrentNote: mockSetCurrentNote,
            isCurrentNote: () => true,
            clearCurrentNote: mockClearCurrentNote,
        });

        render(<NoteCell row={{ original: defaultNoteData }} />);

        const button = screen.getByRole('button');
        fireEvent.click(button);

        expect(mockClearCurrentNote).toHaveBeenCalled();
    });
});
