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
import { vi } from 'vitest';
import { render, screen } from '../../test-utils';
import { SavedQueriesContext } from '../../views';
import ListItemActionMenu from './ListItemActionMenu';
const testSelectedQuery = {
    name: '10 Admins',
    description: '10 Admins desc',
    query: "MATCH p = (t:Group)<-[:MemberOf*1..]-(a)\nWHERE (a:User or a:Computer) and t.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 10",
    canEdit: true,
    id: 1,
    user_id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
};

const TestSavedQueriesContext = {
    selected: { query: 'abcde', id: 1 },
    selectedQuery: testSelectedQuery,
    showSaveQueryDialog: false,
    saveAction: undefined,
    setSelected: () => {},
    setShowSaveQueryDialog: () => {},
    runQuery: vi.fn(),
    editQuery: vi.fn(),
    setSaveAction: () => {},
};

describe('ListItemActionMenu', () => {
    const testDeleteHandler = vi.fn();
    const user = userEvent.setup();
    const ListItemActionMenuWithProvider = () => (
        <SavedQueriesContext.Provider value={TestSavedQueriesContext}>
            <ListItemActionMenu id={1} deleteQuery={testDeleteHandler} query='abc' />
        </SavedQueriesContext.Provider>
    );

    it('renders a ListItemActionMenu component', async () => {
        render(<ListItemActionMenu id={1} deleteQuery={testDeleteHandler} />);

        expect(screen.getByTestId('saved-query-action-menu-trigger')).toBeInTheDocument();
    });

    it('renders the popup content with run, edit/share, and delete when the menu trigger', async () => {
        render(<ListItemActionMenu id={1} deleteQuery={testDeleteHandler} />);

        expect(screen.getByTestId('saved-query-action-menu-trigger')).toBeInTheDocument();

        await user.click(screen.getByRole('button'));
        expect(screen.getByText(/run/i)).toBeInTheDocument();
        expect(screen.getByText(/edit\/share/i)).toBeInTheDocument();
        expect(screen.getByText(/delete/i)).toBeInTheDocument();
    });

    it('fires delete when edit is clicked', async () => {
        render(<ListItemActionMenu id={1} deleteQuery={testDeleteHandler} />);

        expect(screen.getByTestId('saved-query-action-menu-trigger')).toBeInTheDocument();

        await user.click(screen.getByRole('button'));

        await user.click(screen.getByText(/delete/i));
        expect(testDeleteHandler).toBeCalled();
    });

    it('closes', async () => {
        render(<ListItemActionMenu id={1} deleteQuery={testDeleteHandler} />);

        expect(screen.getByTestId('saved-query-action-menu-trigger')).toBeInTheDocument();

        const trigger = screen.getByRole('button');
        await user.click(trigger);
        expect(screen.getByText(/edit\/share/i)).toBeInTheDocument();

        await user.click(trigger);
        expect(screen.queryByText(/edit\/share/i)).not.toBeInTheDocument();
    });

    it('fires runQuery in context provider', async () => {
        render(<ListItemActionMenuWithProvider />);
        await user.click(screen.getByRole('button'));
        const run = screen.getByText(/run/i);
        expect(run).toBeInTheDocument();
        await user.click(run);
        expect(TestSavedQueriesContext.runQuery).toBeCalled();
    });

    it('fires editQuery in context provider', async () => {
        render(<ListItemActionMenuWithProvider />);
        await user.click(screen.getByRole('button'));
        const edit = screen.getByText('Edit/Share');
        expect(edit).toBeInTheDocument();
        await user.click(edit);
        expect(TestSavedQueriesContext.editQuery).toBeCalled();
    });
});
