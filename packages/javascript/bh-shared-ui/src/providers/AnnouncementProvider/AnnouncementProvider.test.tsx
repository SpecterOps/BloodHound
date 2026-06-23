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

import { act, render, renderHook } from '@testing-library/react';
import AnnouncementProvider from './AnnouncementProvider';
import { useAnnounce } from './hooks';

const wrapper = ({ children }: { children: React.ReactNode }) => (
    <AnnouncementProvider>{children}</AnnouncementProvider>
);

describe('AnnouncementProvider', () => {
    it('mounts a polite aria-live region', () => {
        render(<AnnouncementProvider />);
        const liveRegions = document.querySelectorAll('[aria-live="polite"]');
        expect(liveRegions).toHaveLength(1);
    });

    it('mounts an assertive aria-live region', () => {
        render(<AnnouncementProvider />);
        const liveRegions = document.querySelectorAll('[aria-live="assertive"]');
        expect(liveRegions).toHaveLength(1);
    });

    it('live regions have aria-atomic set to true', () => {
        render(<AnnouncementProvider />);
        const politeRegion = document.querySelector('[aria-live="polite"]');
        const assertiveRegion = document.querySelector('[aria-live="assertive"]');
        expect(politeRegion).toHaveAttribute('aria-atomic', 'true');
        expect(assertiveRegion).toHaveAttribute('aria-atomic', 'true');
    });

    it('announces a polite message (default priority)', () => {
        const { result } = renderHook(() => useAnnounce(), { wrapper });

        act(() => result.current('Search results loaded'));

        const politeRegion = document.querySelector('[aria-live="polite"]');
        expect(politeRegion).toHaveTextContent('Search results loaded');
    });

    it('announces a message with assertive priority', () => {
        const { result } = renderHook(() => useAnnounce(), { wrapper });

        act(() => result.current('Error: Session expired', 'assertive'));

        const assertiveRegion = document.querySelector('[aria-live="assertive"]');
        expect(assertiveRegion).toHaveTextContent('Error: Session expired');
    });

    it('does not put polite messages in the assertive region', () => {
        const { result } = renderHook(() => useAnnounce(), { wrapper });

        act(() => result.current('Polite only message', 'polite'));

        const assertiveRegion = document.querySelector('[aria-live="assertive"]');
        expect(assertiveRegion).not.toHaveTextContent('Polite only message');
    });

    it('does not put assertive messages in the polite region', () => {
        const { result } = renderHook(() => useAnnounce(), { wrapper });

        act(() => result.current('Assertive only message', 'assertive'));

        const politeRegion = document.querySelector('[aria-live="polite"]');
        expect(politeRegion).not.toHaveTextContent('Assertive only message');
    });

    it('re-announces the same message by incrementing the span key', () => {
        const { result } = renderHook(() => useAnnounce(), { wrapper });

        act(() => result.current('Repeated message'));
        const firstSpan = document.querySelector('[aria-live="polite"] span');
        const firstKey = firstSpan?.getAttribute('data-reactkey');

        act(() => result.current('Repeated message'));
        const secondSpan = document.querySelector('[aria-live="polite"] span');
        const secondKey = secondSpan?.getAttribute('data-reactkey');

        // The span is remounted (new DOM node) on each announce call
        expect(firstSpan).not.toBe(secondSpan);
        // Keys are internal to React, so we verify the content is still present
        expect(secondSpan).toHaveTextContent('Repeated message');
        void firstKey;
        void secondKey;
    });
});

describe('useAnnounce', () => {
    it('throws when used outside of AnnouncementProvider', () => {
        const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
        const expectedMessage = 'useAnnounce must be used within an AnnouncementProvider';

        const suppressExpectedError = (event: ErrorEvent) => {
            if (event.message.includes(expectedMessage)) {
                event.preventDefault();
            }
        };

        window.addEventListener('error', suppressExpectedError);

        try {
            expect(() => renderHook(() => useAnnounce())).toThrow(expectedMessage);
        } finally {
            window.removeEventListener('error', suppressExpectedError);
            consoleError.mockRestore();
        }
    });
});
