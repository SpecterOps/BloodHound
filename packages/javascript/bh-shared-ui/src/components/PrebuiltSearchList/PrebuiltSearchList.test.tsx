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
            category: 'Active Directory',
            subheader: 'Domain Information',
            queries: [
                {
                    name: 'Find all Domain Admins',
                    description: 'Returns all members of the Domain Admins group',
                    query: 'MATCH (n:Group {name:"DOMAIN ADMINS@TESTLAB.LOCAL"})<-[r:MemberOf*1..]-(m) RETURN m',
                },
                {
                    name: 'Find Computers with Unconstrained Delegation',
                    description: 'Returns all computers with unconstrained delegation enabled',
                    query: 'MATCH (c:Computer {unconstraineddelegation:true}) RETURN c',
                },
                {
                    name: 'Find Kerberoastable Users',
                    description: 'Returns all users with SPN set',
                    query: 'MATCH (u:User {hasspn:true}) RETURN u',
                    canEdit: true,
                    id: 1,
                    user_id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
                },
            ],
        },
        {
            category: 'Azure',
            subheader: 'Tenant Information',
            queries: [
                {
                    name: 'Find all Global Admins',
                    description: 'Returns all members of the Global Administrator role',
                    query: 'MATCH (n:AZRole {name:"Global Administrator"})<-[r:AZHasRole]-(m) RETURN m',
                },
                {
                    name: 'Find Azure VMs',
                    description: 'Returns all Azure virtual machines',
                    query: 'MATCH (v:AZVM) RETURN v',
                },
            ],
        },
    ];

    const setup = () => {
        const testClickHandler = vi.fn();
        const testDeleteHandler = vi.fn();
        const testClearFiltersHandler = vi.fn();

        const result = render(
            <PrebuiltSearchList
                listSections={testListSections}
                showCommonQueries={true}
                clickHandler={testClickHandler}
                deleteHandler={testDeleteHandler}
                clearFiltersHandler={testClearFiltersHandler}
            />
        );

        return {
            ...result,
            testClickHandler,
            testDeleteHandler,
            testClearFiltersHandler,
        };
    };

    it('renders a list of pre-built searches', async () => {
        setup();

        expect(screen.getByText('Domain Information')).toBeInTheDocument();
        expect(screen.getByText('Find all Domain Admins')).toBeInTheDocument();

        // Verifying that the one editable query has an action menu
        const actionMenuTriggers = screen.getAllByTestId('saved-query-action-menu-trigger');
        expect(actionMenuTriggers).toHaveLength(1);
    });

    it('calls clickHandler when a line item is clicked', async () => {
        const user = userEvent.setup();
        const { testClickHandler } = setup();

        await user.click(screen.getByText(testListSections[0].queries[0].name));

        expect(testClickHandler).toBeCalledWith(testListSections[0].queries[0].query, undefined);

        await user.click(screen.getByText(testListSections[0].queries[2].name));

        expect(testClickHandler).toBeCalledWith(testListSections[0].queries[2].query, 1);
    });

    it('clicking a delete button calls deleteHandler', async () => {
        const user = userEvent.setup();
        setup();

        const actionMenuTrigger = screen.getByTestId('saved-query-action-menu-trigger');
        await user.click(actionMenuTrigger);

        const deleteOption = await screen.findByText(/delete/i);
        expect(deleteOption).toBeInTheDocument();
    });

    it('renders a list of pre-built searches and displays categories and subcategories one time', async () => {
        setup();

        const adCategory = screen.getAllByText('Active Directory');
        expect(adCategory).toHaveLength(1);

        const domainInfoSubcategory = screen.getAllByText('Domain Information');
        expect(domainInfoSubcategory).toHaveLength(1);

        const azureCategory = screen.getAllByText('Azure');
        expect(azureCategory).toHaveLength(1);

        const tenantInfoSubcategory = screen.getAllByText('Tenant Information');
        expect(tenantInfoSubcategory).toHaveLength(1);
    });
});
