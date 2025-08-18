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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, render, screen, waitFor } from '../../../test-utils';
import SaveQueryDialog from './SaveQueryDialog';
const testUsers = [
    {
        principal_name: 'Ted Theodore Logan - user',
        id: '8f98c54d-ac4b-464d-a391-d8d5d4b3fe8c',
    },
    {
        principal_name: 'Socrates - read-only',
        id: 'cd28625d-5a09-4312-b84a-72a95881387a',
    },
    {
        principal_name: 'Beethoven - upload-only',
        id: '46cd933d-b556-4fd6-a7a5-c0c7a44ea11e',
    },
    {
        principal_name: 'Joan of Arc - power-user',
        id: 'b1ebcec3-97bd-4660-8a88-b1895cbf4859',
    },
    {
        principal_name: 'Bill S Preston - admin',
        id: '0bf8e936-c70b-4ddc-ad8a-98a3afbf6393',
    },
];

const testSelf = {
    id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
};

const testPermissions = {
    query_id: 3,
    public: true,
    shared_to_user_ids: [],
};

const handlers = [
    rest.get('/api/v2/bloodhound-users', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    users: testUsers,
                },
            })
        );
    }),
    rest.get(`/api/v2/self`, async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: testSelf,
            })
        );
    }),
    rest.get(`/api/v2/saved-queries/:queryId/permissions`, async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: testPermissions,
            })
        );
    }),
    rest.get('/api/v2/graphs/kinds', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: {},
            })
        );
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('SaveQueryDialog', () => {
    it('should render a SaveQueryDialog', () => {
        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testError = undefined;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };

        render(
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                error={testError}
                cypherSearchState={testCypherSearchState}
                selectedQuery={undefined}
                sharedIds={[]}
                isPublic={false}
                saveAction={''}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        expect(screen.getByText(/save query/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/query name/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument();
    });

    it('should disable save button while input is empty', async () => {
        const user = userEvent.setup();

        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testError = undefined;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };

        const testQueryName = 'query name';

        render(
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                error={testError}
                cypherSearchState={testCypherSearchState}
                selectedQuery={undefined}
                sharedIds={[]}
                isPublic={false}
                saveAction={''}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        expect(screen.getByRole('button', { name: /save/i })).toBeDisabled();

        await user.type(screen.getByLabelText(/query name/i), testQueryName);

        expect(screen.getByRole('button', { name: /save/i })).not.toBeDisabled();
    });

    it('should call onClose when cancel button is clicked', async () => {
        const user = userEvent.setup();
        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testError = undefined;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };

        render(
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                error={testError}
                cypherSearchState={testCypherSearchState}
                selectedQuery={undefined}
                sharedIds={[]}
                isPublic={false}
                saveAction={''}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        await user.click(screen.getByRole('button', { name: /cancel/i }));

        expect(testOnClose).toHaveBeenCalled();
    });

    it('should call onSave when save button is clicked', async () => {
        const user = userEvent.setup();
        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testError = undefined;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };
        const testQueryName = 'query name';

        render(
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                error={testError}
                cypherSearchState={testCypherSearchState}
                selectedQuery={undefined}
                sharedIds={[]}
                isPublic={false}
                saveAction={''}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        await user.type(screen.getByLabelText(/query name/i), testQueryName);

        await user.click(screen.getByRole('button', { name: /save/i }));

        expect(testOnSave).toHaveBeenCalled();
    });

    it('should display an error when error prop is truthy', () => {
        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testError = true;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };

        render(
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                error={testError}
                cypherSearchState={testCypherSearchState}
                selectedQuery={undefined}
                sharedIds={[]}
                isPublic={false}
                saveAction={''}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        expect(
            screen.getByText(/an error ocurred while attempting to save this query. please try again./i)
        ).toBeInTheDocument();
    });

    it('should render an SavedQueryPermissions component', async () => {
        const testOnSave = vitest.fn();
        const testOnClose = vitest.fn();
        const testOnUpdate = vitest.fn();
        const testSetSharedIds = vitest.fn();
        const testSetIsPublic = vitest.fn();

        const testError = true;
        const testCypherSearchState = {
            cypherQuery: '',
            performSearch: vitest.fn(),
            setCypherQuery: vitest.fn(),
        };
        const testSelectedQuery = {
            name: 'Uploaded Query 3',
            description: 'Uploaded Query 3 Desc',
            query: "MATCH p = (t:Group)<-[:MemberOf*1..]-(a) WHERE (a:User or a:Computer) and t.objectid ENDS WITH '-512' RETURN p LIMIT 104",
            canEdit: true,
            id: 4,
            user_id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
        };

        await act(async () =>
            render(
                <SaveQueryDialog
                    open
                    onSave={testOnSave}
                    onClose={testOnClose}
                    error={testError}
                    cypherSearchState={testCypherSearchState}
                    selectedQuery={testSelectedQuery}
                    sharedIds={[]}
                    isPublic={false}
                    saveAction={''}
                    onUpdate={testOnUpdate}
                    setSharedIds={testSetSharedIds}
                    setIsPublic={testSetIsPublic}
                />
            )
        );

        // Table Header Rendered
        const nestedElement = await waitFor(() => screen.getByText(/Manage Shared Queries/i));
        expect(nestedElement).toBeInTheDocument();

        const testTable = screen.getByRole('table');
        expect(testTable).toBeInTheDocument();
    });
});
