// Copyright 2024 Specter Ops, Inc.
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

import { render, screen } from 'src/test-utils';
import DatabaseManagement from '.';
import userEvent from '@testing-library/user-event';

describe('DatabaseManagement', () => {
    beforeEach(() => {
        render(<DatabaseManagement />);
    });

    it('renders', () => {
        const title = screen.getByText(/clear bloodhound data/i);
        const checkboxes = screen.getAllByRole('checkbox');
        const button = screen.getByRole('button', { name: /proceed/i });

        expect(title).toBeInTheDocument();
        expect(checkboxes.length).toEqual(4);
        expect(button).toBeInTheDocument();
    });

    it('displays error if proceed button is clicked when no checkbox is selected', async () => {
        const user = userEvent.setup();

        const button = screen.getByRole('button', { name: /proceed/i });
        await user.click(button);

        const errorMsg = screen.getByText(/please make a selection/i);
        expect(errorMsg).toBeInTheDocument();
    });

    it('open and closes dialog', async () => {
        const user = userEvent.setup();

        const checkbox = screen.getByRole('checkbox', { name: /collected graph data/i });
        await user.click(checkbox);

        const button = screen.getByRole('button', { name: /proceed/i });
        await user.click(button);

        const dialog = screen.getByRole('dialog', { name: /confirm deleting data/i });
        expect(dialog).toBeInTheDocument();

        const closeButton = screen.getByRole('button', { name: /cancel/i });
        await user.click(closeButton);

        expect(dialog).not.toBeInTheDocument();
    });

    it('handles posting a mutation', async () => {});
});
