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
import { DocPageContent } from './components/DocPageContent';

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

        expect(cards.map((card) => card.getAttribute('href'))).toEqual(['/doodle-ui/embedded', '/docs']);
    });

    it('renders the DoodleUI Storybook Link Card linking to the embedded view', () => {
        renderAt('/');

        const card = screen.getByRole('link', { name: /DoodleUI Storybook/ });

        expect(card).toHaveAttribute('href', '/doodle-ui/embedded');
        expect(within(card).getByText('Storybook inside the Dev Portal')).toBeInTheDocument();
        expect(card.querySelector('svg[data-icon="window-maximize"]')).toBeInTheDocument();
    });

    it('renders the Repo Docs Link Card linking to the markdown section', () => {
        renderAt('/');

        const card = screen.getByRole('link', { name: /Repo Docs/ });

        expect(card).toHaveAttribute('href', '/docs');
        expect(within(card).getByText('Read RFCs, guides, and READMEs')).toBeInTheDocument();
        expect(card.querySelector('svg[data-icon="book"]')).toBeInTheDocument();
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

describe('Dev Portal Repo Docs', () => {
    // One representative Doc Page per include-list entry, as [folder group, page name, route]
    const representativeDocPages = [
        ['rfc', 'bh-rfc-0', '/docs/rfc/bh-rfc-0'],
        ['Repo root', 'DEVREADME', '/docs/DEVREADME'],
        ['Repo root', 'CONTRIBUTING', '/docs/CONTRIBUTING'],
        ['server', 'README', '/docs/server/README'],
        ['packages/javascript/doodle-ui', 'README', '/docs/packages/javascript/doodle-ui/README'],
    ];

    const sidebar = () => within(screen.getByRole('navigation', { name: 'Doc Pages' }));
    const sidebarGroup = (folder: string) => within(sidebar().getByRole('list', { name: folder }));

    it('opens from the Repo Docs Link Card with a sidebar listing the included Doc Pages grouped by folder', async () => {
        const user = userEvent.setup();
        renderAt('/');

        await user.click(screen.getByRole('link', { name: /Repo Docs/ }));

        representativeDocPages.forEach(([folder, name, route]) => {
            expect(sidebarGroup(folder).getByRole('link', { name })).toHaveAttribute('href', route);
        });
    });

    it('labels Doc Pages by file name without the .md extension', () => {
        renderAt('/docs');

        const linkNames = sidebar()
            .getAllByRole('link')
            .map((link) => link.textContent);
        expect(linkNames.length).toBeGreaterThan(0);
        linkNames.forEach((linkName) => {
            expect(linkName).not.toMatch(/\.md$/);
            expect(linkName).not.toContain('/');
        });
    });

    it('renders a selected Doc Page with GitHub-flavored markdown and highlights it', async () => {
        const user = userEvent.setup();
        renderAt('/docs');

        await user.click(sidebarGroup('rfc').getByRole('link', { name: 'bh-rfc-2' }));

        expect(screen.getByRole('heading', { level: 1, name: 'Conventional Commits Guide' })).toBeInTheDocument();
        expect(screen.getByRole('columnheader', { name: 'Type' })).toBeInTheDocument();
        expect(screen.getByRole('cell', { name: 'feat' })).toBeInTheDocument();
        expect(sidebarGroup('rfc').getByRole('link', { name: 'bh-rfc-2' })).toHaveAttribute('aria-current', 'page');
        expect(sidebarGroup('Repo root').getByRole('link', { name: 'DEVREADME' })).not.toHaveAttribute('aria-current');
    });

    it('renders the Doc Page at a direct address', () => {
        renderAt('/docs/packages/javascript/doodle-ui/README');

        expect(screen.getByRole('heading', { level: 1, name: 'DoodleUI' })).toBeInTheDocument();
        expect(sidebarGroup('packages/javascript/doodle-ui').getByRole('link', { name: 'README' })).toHaveAttribute(
            'aria-current',
            'page'
        );
    });

    it('shows the not-found page for an unknown Doc Page address', () => {
        renderAt('/docs/rfc/bh-rfc-999');

        expect(screen.getByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
    });

    it('offers a way back to the landing page from a Doc Page', async () => {
        const user = userEvent.setup();
        renderAt('/docs/DEVREADME');

        await user.click(screen.getByRole('link', { name: 'Dev Portal' }));

        expect(screen.getByRole('link', { name: /Repo Docs/ })).toBeInTheDocument();
    });

    it('switches to another Link Card destination from the header menu', async () => {
        const user = userEvent.setup();
        renderAt('/docs/DEVREADME');

        await user.click(screen.getByRole('button', { name: /Repo Docs/ }));
        expect(screen.getByRole('menuitem', { name: /Repo Docs/ })).toHaveAttribute('aria-current', 'page');
        await user.click(screen.getByRole('menuitem', { name: /DoodleUI Storybook/ }));

        expect(screen.getByTitle('DoodleUI Storybook')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /DoodleUI Storybook/ })).toBeInTheDocument();
    });

    describe('search', () => {
        const searchResults = () => within(screen.getByRole('list', { name: 'Search results' }));

        beforeEach(() => {
            Element.prototype.scrollIntoView = vi.fn();
        });

        it('opens a results page listing matching Doc Pages when a query is submitted', async () => {
            const user = userEvent.setup();
            renderAt('/docs');

            await user.type(screen.getByRole('searchbox', { name: 'Search Repo Docs' }), 'Conventional Commits{Enter}');

            expect(
                screen.getByRole('heading', { level: 1, name: 'Search results for “Conventional Commits”' })
            ).toBeInTheDocument();
            expect(
                searchResults().getByRole('link', { name: 'bh-rfc-2 › Conventional Commits Guide' })
            ).toHaveAttribute('href', '/docs/rfc/bh-rfc-2#conventional-commits-guide');
            expect(screen.getByRole('searchbox', { name: 'Search Repo Docs' })).toHaveValue('Conventional Commits');
        });

        it('matches partial words', () => {
            renderAt('/docs?q=Conventi');

            expect(
                searchResults().getByRole('link', { name: 'bh-rfc-2 › Conventional Commits Guide' })
            ).toBeInTheDocument();
        });

        it('links a result to its section and scrolls to that heading when opened', async () => {
            const user = userEvent.setup();
            renderAt('/docs?q=RFC Numbering');

            const result = searchResults().getByRole('link', { name: 'bh-rfc-0 › 4. RFC Numbering' });
            expect(result).toHaveAttribute('href', '/docs/rfc/bh-rfc-0#4-rfc-numbering');

            await user.click(result);

            const heading = screen.getByRole('heading', { level: 2, name: '4. RFC Numbering' });
            expect(heading).toHaveAttribute('id', '4-rfc-numbering');
            expect(heading.scrollIntoView).toHaveBeenCalled();
        });

        it('shows a message when nothing matches', () => {
            renderAt('/docs?q=zzqqxxyy');

            expect(screen.getByText('No Doc Pages match your search.')).toBeInTheDocument();
        });

        it('keeps the sidebar folder list on the results page', () => {
            renderAt('/docs?q=Conventional');

            expect(sidebarGroup('rfc').getByRole('link', { name: 'bh-rfc-2' })).toHaveAttribute(
                'href',
                '/docs/rfc/bh-rfc-2'
            );
        });

        it('returns to the Repo Docs start page when an empty query is submitted', async () => {
            const user = userEvent.setup();
            renderAt('/docs?q=Conventional');

            await user.clear(screen.getByRole('searchbox', { name: 'Search Repo Docs' }));
            await user.type(screen.getByRole('searchbox', { name: 'Search Repo Docs' }), '{Enter}');

            expect(screen.getByText('Select a Doc Page from the sidebar.')).toBeInTheDocument();
        });
    });

    it('gives headings GitHub-style ids that in-page links point at', () => {
        renderAt('/docs/rfc/bh-rfc-0');

        expect(screen.getByRole('link', { name: 'section 4' })).toHaveAttribute('href', '#4-rfc-numbering');
        expect(screen.getByRole('heading', { level: 2, name: '4. RFC Numbering' })).toHaveAttribute(
            'id',
            '4-rfc-numbering'
        );
        expect(screen.getByRole('heading', { level: 2, name: '2. Motivation & Goals' })).toHaveAttribute(
            'id',
            '2-motivation--goals'
        );
    });
});

describe('Dev Portal Doc Page links', () => {
    const content = () => within(screen.getByRole('main'));

    it('sends a link to a non-included repo file to GitHub', () => {
        renderAt('/docs/rfc/bh-rfc-7');

        const link = content().getByRole('link', { name: 'bhce/AGENTS.md' });
        expect(link).toHaveAttribute('href', 'https://github.com/SpecterOps/BloodHound/blob/main/AGENTS.md');
        expect(link).toHaveAttribute('target', '_blank');
        expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    });

    it('sends a link to a repo folder to that folder on GitHub', () => {
        renderAt('/docs/server/README');

        expect(content().getByRole('link', { name: 'docs/architecture/' })).toHaveAttribute(
            'href',
            'https://github.com/SpecterOps/BloodHound/tree/main/server/docs/architecture/'
        );
    });

    it('points a link to an included file at its Doc Page', () => {
        renderAt('/docs/packages/javascript/doodle-ui/README');

        const link = content().getByRole('link', { name: 'README' });
        expect(link).toHaveAttribute('href', '/docs/packages/javascript/doodle-ui/README');
        expect(link).not.toHaveAttribute('target');
    });

    it('opens external links in a new tab and leaves in-page anchors in place', () => {
        renderAt('/docs/rfc/bh-rfc-0');

        const externalLink = content().getByRole('link', { name: 'GitHub Flavored Markdown' });
        expect(externalLink).toHaveAttribute(
            'href',
            'https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax'
        );
        expect(externalLink).toHaveAttribute('target', '_blank');
        expect(externalLink).toHaveAttribute('rel', 'noopener noreferrer');
        const anchorLink = content().getByRole('link', { name: 'section 4' });
        expect(anchorLink).toHaveAttribute('href', '#4-rfc-numbering');
        expect(anchorLink).not.toHaveAttribute('target');
        expect(content().getByRole('img', { name: 'xkcd comic - standards' })).toHaveAttribute(
            'src',
            'https://imgs.xkcd.com/comics/standards.png'
        );
    });
});

describe('Doc Page markdown resolution', () => {
    const renderMarkdown = (markdown: string) =>
        render(
            <MemoryRouter>
                <DocPageContent docPage={{ sourcePath: 'rfc/example.md', route: '/docs/rfc/example', markdown }} />
            </MemoryRouter>
        );

    it('leaves protocol-relative links and images unchanged', () => {
        renderMarkdown('[Example](//example.com/page) ![Logo](//example.com/logo.png)');

        expect(screen.getByRole('link', { name: 'Example' })).toHaveAttribute('href', '//example.com/page');
        expect(screen.getByRole('img', { name: 'Logo' })).toHaveAttribute('src', '//example.com/logo.png');
    });

    it('leaves the src of a relative image that is not bundled unchanged', () => {
        renderMarkdown('![Diagram](images/missing-diagram.png)');

        expect(screen.getByRole('img', { name: 'Diagram' })).toHaveAttribute('src', 'images/missing-diagram.png');
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
