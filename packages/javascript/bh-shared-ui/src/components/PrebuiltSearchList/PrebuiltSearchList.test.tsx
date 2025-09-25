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

import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';
import { render, screen } from '../../test-utils';
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
        const testClickHandler = vi.fn();
        const testDeleteHandler = vi.fn();
        const testClearFiltersHandler = vi.fn();

        render(
            <PrebuiltSearchList
                listSections={testListSections}
                showCommonQueries={true}
                clickHandler={testClickHandler}
                deleteHandler={testDeleteHandler}
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
        const testClickHandler = vi.fn();
        const testDeleteHandler = vi.fn();
        const testClearFiltersHandler = vi.fn();

        render(
            <PrebuiltSearchList
                listSections={testListSections}
                showCommonQueries={true}
                clickHandler={testClickHandler}
                deleteHandler={testDeleteHandler}
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
        const testClickHandler = vi.fn();
        const testDeleteHandler = vi.fn();
        const testClearFiltersHandler = vi.fn();

        render(
            <PrebuiltSearchList
                listSections={testListSections}
                showCommonQueries={true}
                clickHandler={testClickHandler}
                deleteHandler={testDeleteHandler}
                clearFiltersHandler={testClearFiltersHandler}
            />
        );

        await user.click(screen.getByRole('button'));
        expect(screen.getByText(/delete/i)).toBeInTheDocument();
    });
});
