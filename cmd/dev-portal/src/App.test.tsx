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

import { act, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, RouterProvider, createMemoryRouter } from 'react-router-dom';
import App from './App';

const renderAt = (url: string) =>
    render(
        <MemoryRouter initialEntries={[url]}>
            <App />
        </MemoryRouter>
    );

describe('Dev Portal landing page', () => {
    it('renders one Link Card per configured destination', () => {
        renderAt('/');

        const cards = within(screen.getByRole('list')).getAllByRole('link');

        expect(cards.map((card) => card.getAttribute('href'))).toEqual(['/doodle-ui/embedded']);
    });

    it('renders the DoodleUI Storybook Link Card linking to the embedded view', () => {
        renderAt('/');

        const card = screen.getByRole('link', { name: /DoodleUI Storybook/ });

        expect(card).toHaveAttribute('href', '/doodle-ui/embedded');
        expect(within(card).getByText('Storybook inside the Dev Portal')).toBeInTheDocument();
        expect(card.querySelector('svg[data-icon="window-maximize"]')).toBeInTheDocument();
    });
});

describe('Dev Portal embedded DoodleUI view', () => {
    it("loads the Storybook build in an iframe at the portal's ?path=", () => {
        renderAt('/doodle-ui/embedded?path=/story/button--primary');

        expect(screen.getByTitle('DoodleUI Storybook')).toHaveAttribute(
            'src',
            '/artifacts/doodle-ui/?path=%2Fstory%2Fbutton--primary'
        );
    });

    it('loads the Storybook build at its default story when the portal has no ?path=', () => {
        renderAt('/doodle-ui/embedded');

        expect(screen.getByTitle('DoodleUI Storybook')).toHaveAttribute('src', '/artifacts/doodle-ui/');
    });

    it('shows the Dev Portal header with a way back to the landing page', async () => {
        const user = userEvent.setup();
        renderAt('/doodle-ui/embedded');

        await user.click(within(screen.getByRole('banner')).getByRole('link', { name: 'Dev Portal' }));

        expect(screen.getByRole('link', { name: /DoodleUI Storybook/ })).toBeInTheDocument();
    });

    it("updates the portal's ?path= with a history push when the Storybook iframe navigates", async () => {
        const router = createMemoryRouter([{ path: '*', element: <App /> }], {
            initialEntries: ['/doodle-ui/embedded?path=/story/button--primary'],
        });
        render(<RouterProvider router={router} />);
        const iframe = screen.getByTitle('DoodleUI Storybook') as HTMLIFrameElement;
        if (!iframe.contentWindow) throw new Error('iframe has no window');

        iframe.contentWindow.history.pushState(null, '', '/artifacts/doodle-ui/?path=/story/badge--default');
        fireEvent.load(iframe);

        expect(router.state.location.pathname).toBe('/doodle-ui/embedded');
        expect(new URLSearchParams(router.state.location.search).get('path')).toBe('/story/badge--default');
        expect(router.state.historyAction).toBe('PUSH');

        await act(() => router.navigate(-1));

        expect(new URLSearchParams(router.state.location.search).get('path')).toBe('/story/button--primary');
    });

    describe('after back/forward', () => {
        beforeEach(() => {
            vi.useFakeTimers();
        });

        afterEach(() => {
            vi.useRealTimers();
        });

        it("keeps the portal's ?path= until the Storybook iframe reaches the requested story", async () => {
            const router = createMemoryRouter([{ path: '*', element: <App /> }], {
                initialEntries: ['/doodle-ui/embedded?path=/story/button--primary'],
            });
            render(<RouterProvider router={router} />);
            const iframe = screen.getByTitle('DoodleUI Storybook') as HTMLIFrameElement;
            const iframeWindow = iframe.contentWindow;
            if (!iframeWindow) throw new Error('iframe has no window');
            const portalStoryPath = () => new URLSearchParams(router.state.location.search).get('path');

            iframeWindow.history.pushState(null, '', '/artifacts/doodle-ui/?path=/story/badge--default');
            fireEvent.load(iframe);
            expect(portalStoryPath()).toBe('/story/badge--default');

            await act(() => router.navigate(-1));
            act(() => {
                vi.advanceTimersByTime(1000);
            });
            fireEvent.load(iframe);

            expect(portalStoryPath()).toBe('/story/button--primary');
            expect(router.state.historyAction).toBe('POP');

            iframeWindow.history.replaceState(null, '', '/artifacts/doodle-ui/?path=/story/button--primary');
            fireEvent.load(iframe);
            iframeWindow.history.pushState(null, '', '/artifacts/doodle-ui/?path=/story/card--default');
            act(() => {
                vi.advanceTimersByTime(1000);
            });

            expect(portalStoryPath()).toBe('/story/card--default');
            expect(router.state.historyAction).toBe('PUSH');
        });
    });
});

describe('Dev Portal not-found page', () => {
    it('shows a not-found message with a way back to the landing page', async () => {
        const user = userEvent.setup();
        renderAt('/nope');

        expect(screen.getByRole('heading', { name: 'Page not found' })).toBeInTheDocument();

        await user.click(screen.getByRole('link', { name: 'Back to the main page' }));

        expect(screen.getByRole('link', { name: /DoodleUI Storybook/ })).toBeInTheDocument();
    });
});
