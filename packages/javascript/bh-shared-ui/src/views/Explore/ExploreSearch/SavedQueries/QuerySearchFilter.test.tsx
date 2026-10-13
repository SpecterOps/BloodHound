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
    const testPlatforms = ['Active Directory', 'Azure', 'Saved Queries', 'Asset Explorer'];

    it('renders the QuerySearchFilter component', async () => {
        render(
            <QuerySearchFilter
                queryFilterHandler={testHandleFilter}
                exportHandler={testHandleExport}
                deleteHandler={testHandleDeleteQuery}
                categories={testCategories}
                platforms={testPlatforms}
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
                platforms={testPlatforms}
                searchTerm={''}
                platform={''}
                categoryFilter={[]}
                source={''}></QuerySearchFilter>
        );

        const platformsSelect = screen.getByLabelText('Platforms');

        expect(platformsSelect).toBeInTheDocument();

        expect(screen.queryByText('All')).not.toBeInTheDocument();

        await user.click(platformsSelect);

        const testPlatformAll = screen.getByText('All');
        const testPlatformAD = screen.getByText('Active Directory');
        const testPlatformAzure = screen.getByText('Azure');
        const testPlatformSavedQueries = screen.getByText('Saved Queries');
        expect(screen.getByText('Asset Explorer')).toBeInTheDocument();

        expect(testPlatformAll).toBeInTheDocument();
        expect(testPlatformAD).toBeInTheDocument();
        expect(testPlatformAzure).toBeInTheDocument();
        expect(testPlatformSavedQueries).toBeInTheDocument();

        await user.click(testPlatformAzure);
        expect(testHandleFilter).toBeCalledTimes(1);
    });

    it('includes extension queries in the source filter', async () => {
        const user = userEvent.setup();

        render(
            <QuerySearchFilter
                queryFilterHandler={testHandleFilter}
                exportHandler={testHandleExport}
                deleteHandler={testHandleDeleteQuery}
                categories={testCategories}
                platforms={testPlatforms}
                searchTerm=''
                platform=''
                categoryFilter={[]}
                source=''
            />
        );

        await user.click(screen.getByLabelText('Source'));
        const extensionOption = screen.getByRole('option', { name: 'Extension' });

        expect(extensionOption).toBeInTheDocument();
        await user.click(extensionOption);
        expect(testHandleFilter).toHaveBeenCalledWith('', '', [], 'extension');
    });

    it('renders with the Export and Delete buttons disabled', async () => {
        render(
            <QuerySearchFilter
                queryFilterHandler={testHandleFilter}
                exportHandler={testHandleExport}
                deleteHandler={testHandleDeleteQuery}
                categories={testCategories}
                platforms={testPlatforms}
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

    it('hides delete for an extension-managed query even when marked editable', () => {
        mockContext.mockReturnValue({ selectedQuery: { id: 3, canEdit: true, schema_extension_id: 42 } });

        render(
            <QuerySearchFilter
                queryFilterHandler={testHandleFilter}
                exportHandler={testHandleExport}
                deleteHandler={testHandleDeleteQuery}
                categories={testCategories}
                platforms={testPlatforms}
                searchTerm=''
                platform=''
                categoryFilter={[]}
                source=''
            />
        );

        expect(screen.queryByRole('button', { name: /delete/i })).not.toBeInTheDocument();
        mockContext.mockReturnValue({ selectedQuery: undefined });
    });
});
