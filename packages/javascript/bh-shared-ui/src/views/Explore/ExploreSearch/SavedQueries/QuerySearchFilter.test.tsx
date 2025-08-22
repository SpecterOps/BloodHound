import userEvent from '@testing-library/user-event';
import { render, screen } from '../../../../test-utils';
import QuerySearchFilter from './QuerySearchFilter';

const mockProvider = vi.fn();
const mockContext = vi.fn();
vi.mock('../../providers', async () => {
    const actual = await vi.importActual('../../providers');
    return {
        ...actual,
        SavedQueriesProvider: () => mockProvider,
        useSavedQueriesContext: () => mockContext,
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

    const testSelectedQuery = {
        name: '10 Admins',
        description: '10 Admins',
        query: "MATCH p = (t:Group)<-[:MemberOf*1..]-(a)\nWHERE (a:User or a:Computer) and t.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 10",
        canEdit: true,
        id: 1,
        user_id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
    };

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

        expect(testPlatforms).toBeInTheDocument;

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

        const testDelete = screen.getByRole('button', { name: 'delete' });
        expect(testDelete).toBeInTheDocument();
        expect(testDelete).toBeDisabled();
    });
});
