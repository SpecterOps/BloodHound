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

        // Radix SelectValue renders the placeholder ("All") inside the trigger immediately,
        // so we cannot assert it's absent before opening. Instead, verify the listbox (dropdown
        // content) is not yet present.
        expect(screen.queryByRole('listbox')).not.toBeInTheDocument();

        await user.click(testPlatforms);

        // After opening, use role="option" to target SelectItem elements specifically,
        // avoiding false matches against the SelectValue placeholder in the trigger.
        const testPlatformAll = screen.getByRole('option', { name: 'All' });
        const testPlatformAD = screen.getByRole('option', { name: 'Active Directory' });
        const testPlatformAzure = screen.getByRole('option', { name: 'Azure' });
        const testPlatformSavedQueries = screen.getByRole('option', { name: 'Saved Queries' });

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
});
