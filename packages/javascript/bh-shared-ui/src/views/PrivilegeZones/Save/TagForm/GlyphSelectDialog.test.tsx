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
import { ANNOUNCEMENT_DEBOUNCE_MS } from '../../../../components/DebouncedAnnouncement';
import { act, fireEvent, render, screen, waitFor, within } from '../../../../test-utils';
import { freeIconsList } from '../../../../utils';
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

    it('exposes the full glyph count and positions despite virtualized rows', () => {
        render(<GlyphSelectDialog selected={undefined} open={true} onCancel={onCancel} onSelect={onSelect} />);

        const glyphs = within(screen.getByRole('list', { name: 'Glyphs' })).getAllByRole('listitem');

        expect(glyphs.length).toBeLessThan(freeIconsList.length);
        glyphs.forEach((glyph, index) => {
            expect(glyph).toHaveAttribute('aria-setsize', String(freeIconsList.length));
            expect(glyph).toHaveAttribute('aria-posinset', String(index + 1));
        });
    });

    it('announces filtered results without moving focus or making the list live', async () => {
        render(<GlyphSelectDialog selected={undefined} open={true} onCancel={onCancel} onSelect={onSelect} />);

        const searchInput = screen.getByRole('textbox', { name: 'Filter glyphs' });
        const status = screen.getByRole('status');
        const list = screen.getByRole('list', { name: 'Glyphs' });

        expect(status).toHaveAttribute('aria-live', 'polite');
        expect(status).toHaveClass('sr-only');
        expect(status).not.toHaveAttribute('tabindex');
        expect(status).toHaveAttribute('aria-atomic', 'true');
        expect(list).not.toHaveAttribute('aria-live');

        await user.type(searchInput, 'LIGHTBULB');

        await waitFor(() => expect(status).toHaveTextContent('1 matching result'), { timeout: 2000 });
        expect(searchInput).toHaveFocus();
        const glyphs = within(list).getAllByRole('listitem');
        expect(glyphs).toHaveLength(1);
        expect(glyphs[0]).toHaveAttribute('aria-setsize', '1');
        expect(glyphs[0]).toHaveAttribute('aria-posinset', '1');

        await user.type(searchInput, 'nonexistent');

        await waitFor(() => expect(status).toHaveTextContent('No results found'), { timeout: 2000 });
        expect(within(list).queryAllByRole('listitem')).toHaveLength(0);
        expect(searchInput).toHaveFocus();

        await user.clear(searchInput);

        await waitFor(() => expect(status).toHaveTextContent(`${freeIconsList.length} matching results`), {
            timeout: 2000,
        });
        expect(searchInput).toHaveFocus();
    });

    it('filters immediately and replaces equal-count announcements after typing pauses', () => {
        vi.useFakeTimers();
        const { unmount } = render(
            <GlyphSelectDialog selected={undefined} open={true} onCancel={onCancel} onSelect={onSelect} />
        );

        try {
            const searchInput = screen.getByRole('textbox', { name: 'Filter glyphs' });
            const status = screen.getByRole('status');
            expect(status).toHaveTextContent(`${freeIconsList.length} matching results`);
            const initialMessage = status.firstChild;

            act(() => searchInput.focus());
            fireEvent.change(searchInput, { target: { value: 'light' } });
            act(() => vi.advanceTimersByTime(300));
            fireEvent.change(searchInput, { target: { value: 'lightbul' } });

            expect(within(screen.getByRole('list', { name: 'Glyphs' })).getAllByRole('listitem')).toHaveLength(1);
            expect(screen.getByRole('button', { name: 'lightbulb-glyph' })).toBeInTheDocument();
            act(() => vi.advanceTimersByTime(ANNOUNCEMENT_DEBOUNCE_MS - 1));
            expect(status.firstChild).toBe(initialMessage);
            act(() => vi.advanceTimersByTime(1));
            expect(status).toHaveTextContent(/^1 matching result$/);
            expect(status.firstChild).not.toBe(initialMessage);

            const previousMessage = status.firstChild;
            fireEvent.change(searchInput, { target: { value: 'lightbulb' } });
            act(() => vi.advanceTimersByTime(ANNOUNCEMENT_DEBOUNCE_MS - 1));
            expect(status.firstChild).toBe(previousMessage);
            act(() => vi.advanceTimersByTime(1));
            expect(status).toHaveTextContent(/^1 matching result$/);
            expect(status.firstChild).not.toBe(previousMessage);
            expect(searchInput).toHaveFocus();
        } finally {
            unmount();
            vi.useRealTimers();
        }
    });

    it('renders with a selected icon', async () => {
        render(<GlyphSelectDialog selected={'lightbulb'} open={true} onCancel={onCancel} onSelect={onSelect} />);

        expect(screen.getByText('Current Selection:')).toBeInTheDocument();
        expect(screen.getAllByText('lightbulb')).toHaveLength(2);
        await user.type(screen.getByRole('textbox', { name: 'Filter glyphs' }), 'lightbulb');
        expect(screen.getByRole('button', { name: 'lightbulb-glyph', pressed: true })).toBeInTheDocument();
    });

    it('exposes only the current glyph as pressed after keyboard selection', async () => {
        render(<GlyphSelectDialog selected='0' open={true} onCancel={onCancel} onSelect={onSelect} />);

        const list = within(screen.getByRole('list', { name: 'Glyphs' }));
        const selectedGlyph = list.getByRole('button', { name: '0-glyph', pressed: true });
        const otherGlyph = list.getByRole('button', { name: '1-glyph', pressed: false });
        otherGlyph.focus();
        await user.keyboard('{Enter}');

        expect(otherGlyph).toHaveFocus();
        expect(list.getAllByRole('button', { pressed: true })).toEqual([otherGlyph]);
        expect(selectedGlyph).toHaveAttribute('aria-pressed', 'false');
    });

    it('clears the selected icon when clicking the card', async () => {
        render(<GlyphSelectDialog selected='0' open={true} onCancel={onCancel} onSelect={onSelect} />);

        await user.click(screen.getByRole('button', { name: 'Clear selection' }));

        expect(screen.getByText('None Selected')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: '0-glyph', pressed: false })).toBeInTheDocument();
    });

    it('focuses the search input without opening the clear-selection tooltip when reopened', async () => {
        const { rerender } = render(
            <GlyphSelectDialog selected={'lightbulb'} open={true} onCancel={onCancel} onSelect={onSelect} />
        );

        const searchInput = screen.getByPlaceholderText('Search');

        await waitFor(() => expect(document.activeElement).toBe(searchInput));
        expect(screen.queryByRole('tooltip')).toBeNull();

        rerender(<GlyphSelectDialog selected={'lightbulb'} open={false} onCancel={onCancel} onSelect={onSelect} />);
        rerender(<GlyphSelectDialog selected={'lightbulb'} open={true} onCancel={onCancel} onSelect={onSelect} />);

        const reopenedSearchInput = await screen.findByPlaceholderText('Search');
        await waitFor(() => expect(document.activeElement).toBe(reopenedSearchInput));
        expect(screen.queryByRole('tooltip')).toBeNull();
    });

    it('displays a tooltip for clearing the selected icon', async () => {
        render(<GlyphSelectDialog selected={'lightbulb'} open={true} onCancel={onCancel} onSelect={onSelect} />);

        const clearSelectionButton = screen.getByRole('button', { name: 'Clear selection' });

        await user.hover(clearSelectionButton);

        expect(await screen.findByRole('tooltip', { name: 'Clear selection' })).toBeVisible();
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
