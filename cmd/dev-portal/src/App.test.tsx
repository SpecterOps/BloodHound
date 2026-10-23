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

import { render, screen, within } from '@testing-library/react';
import App from './App';

const destinations = [
    { title: 'Doodle UI Storybook', href: '#storybook', accent: 'blue' },
    { title: 'Documentation', href: '#documentation', accent: 'orange' },
];

describe('App', () => {
    beforeEach(() => {
        render(<App />);
    });

    it('renders a single page-level heading for the portal', () => {
        const headings = screen.getAllByRole('heading', { level: 1 });

        expect(headings).toHaveLength(1);
        expect(headings[0]).toHaveTextContent(/BloodHound\s*Developer Portal/);
        expect(screen.getByRole('region', { name: /Developer Portal/ })).toContainElement(headings[0]);
    });

    it('links the brand back to the top of the page', () => {
        const brandLink = screen.getByRole('link', { name: 'BloodHound Developer Portal home' });

        expect(brandLink).toHaveAttribute('href', '#top');
        expect(document.getElementById('top')).toBeInTheDocument();
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
