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
import { render, screen } from '../../../../test-utils';
import QuerySearchFilter from './QuerySearchFilter';

// const mockProvider = vi.fn();
const mockContext = vi.fn().mockReturnValue({
    // Provide only what the component reads; adjust as needed
    selected: undefined,
    selectedQuery: undefined,
    setSelected: vi.fn(),
    runQuery: vi.fn(),
    editQuery: vi.fn(),
} as any);
vi.mock('../../providers', async () => {
    const actual = await vi.importActual('../../providers');
    return {
        ...actual,
        SavedQueriesProvider: actual.SavedQueriesProvider,
        useSavedQueriesContext: () => mockContext(),
    };
});

describe('QuerySearchFilter', () => {
    const testHandleFilter = vi.fn();
    const testHandleExport = vi.fn();
    const testHandleDeleteQuery = vi.fn();

    const testCategories = [
        'Active Directory Certificate Services',
        'Active Directory Hygiene',
        'Azure Hygiene',
        'Cross Platform Attack Paths',
        'Dangerous Privileges',
        'Domain Information',
        'General',
        'Kerberos Interaction',
        'Microsoft Graph',
        'NTLM Relay Attacks',
        'Shortest Paths',
    ];

    it('renders the QuerySearchFilter component', async () => {
        render(
            <QuerySearchFilter
                queryFilterHandler={testHandleFilter}
                exportHandler={testHandleExport}
                deleteHandler={testHandleDeleteQuery}
                categories={testCategories}
                searchTerm={''}
                platform={''}
                categoryFilter={[]}
                source={''}></QuerySearchFilter>
        );

        const testSearch = screen.getByPlaceholderText('Search');
        expect(testSearch).toBeInTheDocument();
    });

    it('renders the Platforms dropdown and handles click event', async () => {
        const user = userEvent.setup();

        render(
            <QuerySearchFilter
                queryFilterHandler={testHandleFilter}
                exportHandler={testHandleExport}
                deleteHandler={testHandleDeleteQuery}
                categories={testCategories}
                searchTerm={''}
                platform={''}
                categoryFilter={[]}
                source={''}></QuerySearchFilter>
        );

        const testPlatforms = screen.getByLabelText('Platforms');

        expect(testPlatforms).toBeInTheDocument();

        expect(screen.queryByText('All')).not.toBeInTheDocument();

        await user.click(testPlatforms);

        const testPlatformAll = screen.getByText('All');
        const testPlatformAD = screen.getByText('Active Directory');
        const testPlatformAzure = screen.getByText('Azure');
        const testPlatformSavedQueries = screen.getByText('Saved Queries');

        expect(testPlatformAll).toBeInTheDocument();
        expect(testPlatformAD).toBeInTheDocument();
        expect(testPlatformAzure).toBeInTheDocument();
        expect(testPlatformSavedQueries).toBeInTheDocument();

        await user.click(testPlatformAzure);
        expect(testHandleFilter).toBeCalledTimes(1);
    });

    it('renders with the Export and Delete buttons disabled', async () => {
        render(
            <QuerySearchFilter
                queryFilterHandler={testHandleFilter}
                exportHandler={testHandleExport}
                deleteHandler={testHandleDeleteQuery}
                categories={testCategories}
                searchTerm={''}
                platform={''}
                categoryFilter={[]}
                source={''}></QuerySearchFilter>
        );

        const testImport = screen.getByText('Import');
        expect(testImport).toBeInTheDocument();

        const testExport = screen.getByText('Export');
        expect(testExport).toBeInTheDocument();
        expect(testExport).toBeDisabled();

        const testDelete = screen.getByRole('button', { name: /delete/i });
        expect(testDelete).toBeInTheDocument();
        expect(testDelete).toBeDisabled();
    });

    it('does a thing', async () => {
        const user = userEvent.setup();
        render(
            <QuerySearchFilter
                queryFilterHandler={testHandleFilter}
                exportHandler={testHandleExport}
                deleteHandler={testHandleDeleteQuery}
                categories={testCategories}
                searchTerm={''}
                platform={''}
                categoryFilter={[]}
                source={''}></QuerySearchFilter>
        );
        const searchInput = screen.getByPlaceholderText(/search/i);
        expect(searchInput).toBeInTheDocument();
        await user.type(searchInput, 'abc');
        expect(testHandleFilter).toBeCalledTimes(3);
    });
});
