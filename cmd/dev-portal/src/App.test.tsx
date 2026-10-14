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
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
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

        const cards = within(screen.getByRole('list')).queryAllByRole('link');

        expect(cards.map((card) => card.getAttribute('href'))).toEqual([]);
    });
});

describe('Dev Portal not-found page', () => {
    it('shows a not-found message with a way back to the landing page', async () => {
        const user = userEvent.setup();
        renderAt('/nope');

        expect(screen.getByRole('heading', { name: 'Page not found' })).toBeInTheDocument();

        await user.click(screen.getByRole('link', { name: 'Back to the main page' }));

        expect(screen.getByRole('heading', { level: 1, name: 'BloodHound Developer Portal' })).toBeInTheDocument();
    });
});
