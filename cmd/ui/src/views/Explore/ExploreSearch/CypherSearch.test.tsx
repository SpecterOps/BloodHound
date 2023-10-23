// Copyright 2023 Specter Ops, Inc.
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

import { act, render, screen } from 'src/test-utils';
import CypherSearch from './CypherSearch';
import userEvent from '@testing-library/user-event';

describe('CypherSearch', () => {
    beforeEach(async () => {
        await act(async () => {
            render(<CypherSearch />);
        });
    });
    const user = userEvent.setup();

    it('should render', () => {
        expect(screen.getByText(/cypher search/i)).toBeInTheDocument();

        expect(screen.getByRole('link', { name: /help/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /search/i })).toBeInTheDocument();
    });

    it('should show common cypher searches when user clicks on folder button', async () => {
        const prebuiltSearches = screen.getByText(/pre-built searches/i);
        expect(prebuiltSearches).not.toBeVisible();

        const menu = screen.getByRole('button', { name: /show\/hide saved queries/i });

        await user.click(menu);
        expect(prebuiltSearches).toBeVisible();
    });
});
