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
import { render, screen } from '../../../../test-utils';
import GlyphSelectDialog from './GlyphSelectDialog';

const onCancel = vi.fn();
const onSelect = vi.fn();

describe('Glyph Select Dialog', () => {
    const user = userEvent.setup();

    it('renders', async () => {
        render(<GlyphSelectDialog selected={undefined} open={true} onCancel={onCancel} onSelect={onSelect} />);

        expect(await screen.findByText('Select a Glyph')).toBeInTheDocument();
        expect(
            screen.getByText(
                'The selected glyph will apply to all nodes tagged in this Zone for displaying in the Explore graph.'
            )
        ).toBeInTheDocument();
        expect(screen.getByText('Current Selection:')).toBeInTheDocument();

        expect(screen.getByPlaceholderText('Search')).toBeInTheDocument();

        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Confirm/ })).toBeInTheDocument();
    });

    it('renders with a selected icon', async () => {
        render(<GlyphSelectDialog selected={'lightbulb'} open={true} onCancel={onCancel} onSelect={onSelect} />);

        expect(screen.getByText('Current Selection:')).toBeInTheDocument();
        expect(screen.getAllByText('lightbulb')).toHaveLength(2);
    });

    it('calls the passed in cancel handler when clicking the Cancel button', async () => {
        render(<GlyphSelectDialog selected={'lightbulb'} open={true} onCancel={onCancel} onSelect={onSelect} />);

        await user.click(screen.getByRole('button', { name: /Cancel/ }));

        expect(onCancel).toHaveBeenCalled();
    });

    it('calls the passed in select handler when clicking the Confirm button', async () => {
        render(<GlyphSelectDialog selected={'lightbulb'} open={true} onCancel={onCancel} onSelect={onSelect} />);

        await user.click(screen.getByRole('button', { name: /Confirm/ }));

        expect(onSelect).toHaveBeenCalled();
    });
});
