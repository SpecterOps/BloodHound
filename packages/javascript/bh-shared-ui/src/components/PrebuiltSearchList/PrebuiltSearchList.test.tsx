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

import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import PrebuiltSearchList from './PrebuiltSearchList';

describe('PrebuiltSearchList', () => {
    const testListSections = [
        {
            category: 'category',
            subheader: 'subheader text',
            queries: [
                {
                    name: 'query 1',
                    description: 'query 1 description',
                    query: 'match (n) return n limit 5',
                },
                {
                    name: 'query 2',
                    description: 'query 2 description',
                    query: 'match (n) return n limit 5',
                },
                {
                    name: 'query 3',
                    description: 'query 3  description',
                    query: 'match (n) return n limit 5',
                    canEdit: true,
                    id: 1,
                    user_id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
                },
            ],
        },
    ];

    it('renders a list of pre-built searches', async () => {
        const testClickHandler = vitest.fn();
        const testDeleteHandler = vitest.fn();
        const testEditHandler = vitest.fn();
        const testClearFiltersHandler = vitest.fn();

        render(
            <PrebuiltSearchList
                listSections={testListSections}
                selectedQuery={undefined}
                showCommonQueries={true}
                clickHandler={testClickHandler}
                deleteHandler={testDeleteHandler}
                editHandler={testEditHandler}
                clearFiltersHandler={testClearFiltersHandler}
            />
        );
        expect(screen.getAllByText(/subheader/i)[0]).toBeInTheDocument();
        expect(screen.getByText(/query 1/i)).toBeInTheDocument();

        expect(screen.getByText(testListSections[0].queries[0].name)).toBeInTheDocument();
        expect(screen.getByRole('button')).toHaveAttribute('aria-haspopup');
    });

    it('calls clickHandler when a line item is clicked', async () => {
        const user = userEvent.setup();

        const testClickHandler = vitest.fn();
        const testDeleteHandler = vitest.fn();
        const testEditHandler = vitest.fn();
        const testClearFiltersHandler = vitest.fn();

        render(
            <PrebuiltSearchList
                listSections={testListSections}
                selectedQuery={undefined}
                showCommonQueries={true}
                clickHandler={testClickHandler}
                deleteHandler={testDeleteHandler}
                editHandler={testEditHandler}
                clearFiltersHandler={testClearFiltersHandler}
            />
        );

        await user.click(screen.getByText(testListSections[0].queries[0].name));

        expect(testClickHandler).toBeCalledWith(testListSections[0].queries[0].query, undefined);

        await user.click(screen.getByText(testListSections[0].queries[2].name));

        expect(testClickHandler).toBeCalledWith(testListSections[0].queries[2].query, 1);
    });

    it('clicking a delete button calls deleteHandler', async () => {
        const user = userEvent.setup();

        const testClickHandler = vitest.fn();
        const testDeleteHandler = vitest.fn();
        const testEditHandler = vitest.fn();
        const testClearFiltersHandler = vitest.fn();

        render(
            <PrebuiltSearchList
                listSections={testListSections}
                selectedQuery={undefined}
                showCommonQueries={true}
                clickHandler={testClickHandler}
                deleteHandler={testDeleteHandler}
                editHandler={testEditHandler}
                clearFiltersHandler={testClearFiltersHandler}
            />
        );

        await user.click(screen.getByRole('button'));
        expect(screen.getByText(/delete/i)).toBeInTheDocument();

        await user.click(screen.getByText(/delete/i));
        expect(await screen.findByText(/are you sure you want to delete this query/i)).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: /confirm/i }));
        expect(testDeleteHandler).toBeCalledWith(1);
    });

    it('clicking the run button calls run', async () => {
        const user = userEvent.setup();

        const testClickHandler = vitest.fn();
        const testDeleteHandler = vitest.fn();
        const testEditHandler = vitest.fn();
        const testClearFiltersHandler = vitest.fn();

        render(
            <PrebuiltSearchList
                listSections={testListSections}
                selectedQuery={undefined}
                showCommonQueries={true}
                clickHandler={testClickHandler}
                deleteHandler={testDeleteHandler}
                editHandler={testEditHandler}
                clearFiltersHandler={testClearFiltersHandler}
            />
        );

        const actionMenuTrigger = screen.getByTestId('saved-query-action-menu-trigger');

        expect(actionMenuTrigger).toHaveAttribute('aria-haspopup');
        await user.click(actionMenuTrigger);

        const container = screen.getByTestId('saved-query-action-menu', { exact: true });
        expect(screen.getByTestId('saved-query-action-menu')).toBeInTheDocument();

        const runButton = await within(container).findByText(/run/i);
        await user.click(runButton);

        expect(testClickHandler).toBeCalled();
    });

    it('clicking a edit/share button calls editHandler', async () => {
        const user = userEvent.setup();

        const testClickHandler = vitest.fn();
        const testDeleteHandler = vitest.fn();
        const testEditHandler = vitest.fn();
        const testClearFiltersHandler = vitest.fn();

        render(
            <PrebuiltSearchList
                listSections={testListSections}
                selectedQuery={undefined}
                showCommonQueries={true}
                clickHandler={testClickHandler}
                deleteHandler={testDeleteHandler}
                editHandler={testEditHandler}
                clearFiltersHandler={testClearFiltersHandler}
            />
        );

        await user.click(screen.getByRole('button'));
        expect(screen.getByText(/edit\/share/i)).toBeInTheDocument();

        await user.click(screen.getByText(/edit\/share/i));
        expect(testEditHandler).toBeCalled();
    });
});
