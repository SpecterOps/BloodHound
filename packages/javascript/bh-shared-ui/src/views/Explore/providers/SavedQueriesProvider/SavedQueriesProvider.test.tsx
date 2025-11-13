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
import { render, screen } from '../../../../test-utils';
import { SavedQueriesContext, useSavedQueriesContext } from './SavedQueriesContext';
const testSelectedQuery = {
    name: '10 Admins',
    description: '10 Admins desc',
    query: "MATCH p = (t:Group)<-[:MemberOf*1..]-(a)\nWHERE (a:User or a:Computer) and t.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 10",
    canEdit: true,
    id: 1,
    user_id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
};

const TestSavedQueriesContext = {
    selectedId: 1,
    selectedQuery: testSelectedQuery,
    showSaveQueryDialog: false,
    saveAction: undefined,
    setSelectedId: () => {},
    setShowSaveQueryDialog: () => {},
    runQuery: () => {},
    editQuery: () => {},
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
