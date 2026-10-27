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

import { act, render, screen, within } from '@testing-library/react';
import { HashRouter } from 'react-router-dom';
import App from './App';

const destinations = [
    { title: 'Doodle UI Storybook', href: '#/storybook?path=%2Fdocs%2Fabout--docs', accent: 'blue' },
    { title: 'Documentation', href: '#documentation', accent: 'orange' },
];

const renderAt = (hash: string) => {
    window.location.hash = hash;
    render(
        <HashRouter>
            <App />
        </HashRouter>
    );
};

describe('Landing page', () => {
    beforeEach(() => {
        renderAt('#/');
    });

    it('renders a single page-level heading for the portal', () => {
        const headings = screen.getAllByRole('heading', { level: 1 });

        expect(headings).toHaveLength(1);
        expect(headings[0]).toHaveTextContent(/BloodHound\s*Developer Portal/);
        expect(screen.getByRole('region', { name: /Developer Portal/ })).toContainElement(headings[0]);
    });

    it('links the brand to the landing page', () => {
        const brandLink = screen.getByRole('link', { name: 'BloodHound Developer Portal home' });

        expect(brandLink).toHaveAttribute('href', '#/');
    });

    it('renders a portal card for each destination', () => {
        const cardGrid = screen.getByRole('region', { name: 'Developer portal destinations' });
        const cards = within(cardGrid).getAllByRole('link');

        expect(cards).toHaveLength(destinations.length);
        destinations.forEach(({ title, accent }, index) => {
            const card = within(cardGrid).getByRole('link', { name: `Open ${title}` });

            expect(card).toBe(cards[index]);
            expect(card).toHaveAttribute('data-accent', accent);
            expect(within(card).getByRole('heading', { level: 2, name: title })).toBeInTheDocument();
        });
    });

    it.each(destinations)('links the $title card to its destination', ({ title, href }) => {
        expect(screen.getByRole('link', { name: `Open ${title}` })).toHaveAttribute('href', href);
    });
});

describe('Storybook page', () => {
    it('embeds the Doodle UI Storybook build', () => {
        renderAt('#/storybook');

        expect(screen.getByTitle('Doodle UI Storybook')).toHaveAttribute('src', 'doodle-ui-storybook/');
    });

    it('links the header back to the landing page', () => {
        renderAt('#/storybook');
        const banner = screen.getByRole('banner');

        expect(within(banner).getByRole('link', { name: 'Developer Portal' })).toHaveAttribute('href', '#/');
        expect(within(banner).getByText('Internal resources')).toBeInTheDocument();
    });

    it('opens the story named in the portal URL', () => {
        renderAt('#/storybook?path=/story/button--primary');

        expect(screen.getByTitle('Doodle UI Storybook')).toHaveAttribute(
            'src',
            'doodle-ui-storybook/?path=%2Fstory%2Fbutton--primary'
        );
    });

    describe('when a different story is selected inside Storybook', () => {
        beforeEach(() => {
            vi.useFakeTimers();
        });

        afterEach(() => {
            vi.useRealTimers();
        });

        it('copies the story into the portal URL', () => {
            renderAt('#/storybook?path=/story/button--primary');
            const iframeSearch = '?path=/story/card--default';
            Object.defineProperty(screen.getByTitle('Doodle UI Storybook'), 'contentWindow', {
                value: {
                    location: { href: `http://localhost/doodle-ui-storybook/${iframeSearch}`, search: iframeSearch },
                },
            });

            act(() => {
                vi.advanceTimersByTime(500);
            });

            expect(window.location.hash).toBe('#/storybook?path=%2Fstory%2Fcard--default');
            expect(screen.getByTitle('Doodle UI Storybook')).toHaveAttribute(
                'src',
                'doodle-ui-storybook/?path=%2Fstory%2Fbutton--primary'
            );
        });
    });
});

describe('Not found page', () => {
    beforeEach(() => {
        renderAt('#/does-not-exist');
    });

    it('renders a not found message with a link back to the landing page', () => {
        expect(screen.getByRole('heading', { level: 1, name: 'Page not found' })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Back to the Developer Portal' })).toHaveAttribute('href', '#/');
        expect(within(screen.getByRole('banner')).getByRole('link', { name: 'Developer Portal' })).toBeInTheDocument();
    });
});
