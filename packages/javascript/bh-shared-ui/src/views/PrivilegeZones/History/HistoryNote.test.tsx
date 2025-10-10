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
import { render, screen } from '../../../test-utils';
import HistoryNote from './HistoryNote';
import { useHistoryTableContext } from './HistoryTableContext';

vi.mock('./HistoryTableContext', () => ({
    useHistoryTableContext: vi.fn(),
}));

vi.mock('../../../components', () => ({
    AppIcon: {
        LinedPaper: (props: any) => <svg data-testid='lined-paper-icon' {...props} />,
    },
}));

describe('HistoryNotes Component', () => {
    it('renders note header correctly', () => {
        // Mock with no note
        (useHistoryTableContext as jest.Mock).mockReturnValue({ currentNote: null });

        render(<HistoryNote />);

        expect(screen.getByText('Note')).toBeInTheDocument();
        expect(screen.getByTestId('lined-paper-icon')).toBeInTheDocument();
    });

    it('does not render note content when currentNote is null', () => {
        (useHistoryTableContext as jest.Mock).mockReturnValue({ currentNote: null });

        render(<HistoryNote />);

        // Should only render the header
        expect(screen.queryByText(/By /)).not.toBeInTheDocument();
    });

    it('renders note content when currentNote is present', () => {
        const mockNote = {
            note: 'Fixed the critical issue',
            createdBy: 'Jane Doe',
            timestamp: '2025-10-08 10:30AM',
        };

        (useHistoryTableContext as jest.Mock).mockReturnValue({ currentNote: mockNote });

        render(<HistoryNote />);

        expect(screen.getByText('Fixed the critical issue')).toBeInTheDocument();
        expect(screen.getByText(/By Jane Doe on 2025-10-08 10:30AM/)).toBeInTheDocument();
    });
});
