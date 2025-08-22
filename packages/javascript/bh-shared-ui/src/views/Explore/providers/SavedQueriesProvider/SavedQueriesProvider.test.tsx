import { render, screen } from '../../../../test-utils';
import { SavedQueriesContext, useSavedQueriesContext } from './SavedQueriesProvider';

const testSelectedQuery = {
    name: '10 Admins',
    description: '10 Admins desc',
    query: "MATCH p = (t:Group)<-[:MemberOf*1..]-(a)\nWHERE (a:User or a:Computer) and t.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 10",
    canEdit: true,
    id: 1,
    user_id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
};

const TestSavedQueriesContext = {
    selected: { query: '', id: 1 },
    selectedQuery: testSelectedQuery,
    showSaveQueryDialog: false,
    saveAction: undefined,
    setSelected: () => {},
    setShowSaveQueryDialog: () => {},
    runQuery: (query: string, id: number) => {},
    editQuery: (id: number) => {},
    setSaveAction: () => {},
};

const TestingComponent = () => {
    const { selectedQuery } = useSavedQueriesContext();
    return (
        <>
            <p data-testid='name'>{selectedQuery?.name}</p>
            <p data-testid='description'>{selectedQuery?.description}</p>
            <p data-testid='query'>{selectedQuery?.query}</p>
        </>
    );
};

describe('SavedQueriesProvider', () => {
    it('passes data to testing component', () => {
        render(
            <SavedQueriesContext.Provider value={TestSavedQueriesContext}>
                <TestingComponent />
            </SavedQueriesContext.Provider>
        );
        const name = screen.getByTestId('name');
        const desc = screen.getByTestId('description');
        const query = screen.getByTestId('query');
        expect(name).toBeInTheDocument();
        expect(desc).toBeInTheDocument();
        expect(name.textContent).toEqual(testSelectedQuery.name);
        expect(desc.textContent).toEqual(testSelectedQuery.description);
        expect(query.textContent).toEqual(testSelectedQuery.query);
    });
});
