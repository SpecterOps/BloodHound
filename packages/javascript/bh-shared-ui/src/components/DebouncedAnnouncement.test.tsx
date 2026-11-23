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
import { act, cleanup, render, screen } from '@testing-library/react';
import { ANNOUNCEMENT_DEBOUNCE_MS, DebouncedAnnouncement } from './DebouncedAnnouncement';

describe('DebouncedAnnouncement', () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => {
        cleanup();
        vi.useRealTimers();
    });

    it('keeps the region mounted and replaces identical messages for new queries', () => {
        const { rerender } = render(<DebouncedAnnouncement announcementKey='' message='10 matching results' />);
        const status = screen.getByRole('status');
        const initialMessage = status.firstChild;

        expect(status).toHaveClass('sr-only');
        expect(status).not.toHaveAttribute('tabindex');
        expect(status).toHaveAttribute('aria-live', 'polite');
        expect(status).toHaveAttribute('aria-atomic', 'true');

        rerender(<DebouncedAnnouncement announcementKey='a' message='5 matching results' />);
        act(() => vi.advanceTimersByTime(300));
        rerender(<DebouncedAnnouncement announcementKey='ab' message='5 matching results' />);
        act(() => vi.advanceTimersByTime(ANNOUNCEMENT_DEBOUNCE_MS - 1));
        expect(status.firstChild).toBe(initialMessage);
        act(() => vi.advanceTimersByTime(1));
        expect(screen.getByRole('status')).toBe(status);
        expect(status).toHaveTextContent(/^5 matching results$/);

        const previousMessage = status.firstChild;
        rerender(<DebouncedAnnouncement announcementKey='abc' message='5 matching results' />);
        act(() => vi.advanceTimersByTime(ANNOUNCEMENT_DEBOUNCE_MS));
        expect(status.firstChild).not.toBe(previousMessage);
        expect(status).toHaveTextContent(/^5 matching results$/);
    });

    it('does not restart the delay for unrelated renders', () => {
        const { rerender } = render(<DebouncedAnnouncement announcementKey='' message='10 matching results' />);
        rerender(<DebouncedAnnouncement announcementKey='a' message='1 matching result' />);
        act(() => vi.advanceTimersByTime(500));
        rerender(<DebouncedAnnouncement announcementKey='a' message='1 matching result' />);
        act(() => vi.advanceTimersByTime(ANNOUNCEMENT_DEBOUNCE_MS - 500));
        expect(screen.getByRole('status')).toHaveTextContent(/^1 matching result$/);
    });

    it('suppresses announcements while disabled and debounces loaded results', () => {
        const { rerender } = render(<DebouncedAnnouncement announcementKey='' message='10 matching results' />);
        const status = screen.getByRole('status');
        rerender(<DebouncedAnnouncement announcementKey='a' message='No results found' enabled={false} />);
        expect(status).toBeEmptyDOMElement();
        act(() => vi.advanceTimersByTime(ANNOUNCEMENT_DEBOUNCE_MS));
        expect(status).toBeEmptyDOMElement();
        rerender(<DebouncedAnnouncement announcementKey='a' message='No results found' />);
        expect(status).toBeEmptyDOMElement();
        act(() => vi.advanceTimersByTime(ANNOUNCEMENT_DEBOUNCE_MS));
        expect(status).toHaveTextContent(/^No results found$/);
    });

    it('cancels pending announcements when unmounted', () => {
        const { rerender, unmount } = render(
            <DebouncedAnnouncement announcementKey='' message='10 matching results' />
        );
        rerender(<DebouncedAnnouncement announcementKey='a' message='No results found' />);
        expect(vi.getTimerCount()).toBeGreaterThan(0);
        unmount();
        expect(vi.getTimerCount()).toBe(0);
    });
});
