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

import PrebuiltSearchList from '../PrebuiltSearchList';
import { screen, render } from '../../test-utils';
import userEvent from '@testing-library/user-event';

describe('PrebuiltSearchList', () => {
    it('renders a list of pre-built searches', () => {
        const testListSections = [
            {
                subheader: 'subheader',
                lineItems: [
                    {
                        id: 1,
                        description: 'query 1',
                        cypher: 'match (n) return n limit 5',
                        canEdit: false,
                    },
                    {
                        id: 2,
                        description: 'query 2',
                        cypher: 'match (n) return n limit 5',
                        canEdit: false,
                    },
                    {
                        id: 3,
                        description: 'query 3',
                        cypher: 'match (n) return n limit 5',
                        canEdit: false,
                    },
                ],
            },
        ];
        const testClickHandler = vitest.fn();

        render(<PrebuiltSearchList listSections={testListSections} clickHandler={testClickHandler} />);

        expect(screen.getByText(/subheader/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: testListSections[0].lineItems[0].description })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: testListSections[0].lineItems[1].description })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: testListSections[0].lineItems[2].description })).toBeInTheDocument();
    });

    it('clicking a pre-built search calls clickHandler', async () => {
        const user = userEvent.setup();
        const testListSections = [
            {
                subheader: 'subheader',
                lineItems: [
                    {
                        id: 1,
                        description: 'query 1',
                        cypher: 'cypher 1',
                        canEdit: false,
                    },
                    {
                        id: 2,
                        description: 'query 2',
                        cypher: 'cypher 2',
                        canEdit: false,
                    },
                    {
                        id: 3,
                        description: 'query 3',
                        cypher: 'cypher 3',
                        canEdit: false,
                    },
                ],
            },
        ];
        const testClickHandler = vitest.fn();

        render(<PrebuiltSearchList listSections={testListSections} clickHandler={testClickHandler} />);

        await user.click(screen.getByText(testListSections[0].lineItems[0].description));
        expect(testClickHandler).toBeCalledWith(testListSections[0].lineItems[0].cypher);

        await user.click(screen.getByText(testListSections[0].lineItems[1].description));
        expect(testClickHandler).toBeCalledWith(testListSections[0].lineItems[1].cypher);

        await user.click(screen.getByText(testListSections[0].lineItems[2].description));
        expect(testClickHandler).toBeCalledWith(testListSections[0].lineItems[2].cypher);
    });

    it('clicking a delete button calls deleteHandler', async () => {
        const user = userEvent.setup();
        const testListSections = [
            {
                subheader: 'subheader',
                lineItems: [
                    {
                        id: 1,
                        description: 'query 1',
                        cypher: 'cypher 1',
                        canEdit: true,
                    },
                ],
            },
        ];
        const testClickHandler = vitest.fn();
        const testDeleteHandler = vitest.fn();

        render(
            <PrebuiltSearchList
                listSections={testListSections}
                clickHandler={testClickHandler}
                deleteHandler={testDeleteHandler}
            />
        );

        await user.click(
            screen.getByRole('button', {
                name: /delete query/i,
            })
        );
        expect(await screen.findByText(/are you sure you want to delete this query/i)).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: /confirm/i }));
        expect(testDeleteHandler).toBeCalledWith(1);
    });
});
