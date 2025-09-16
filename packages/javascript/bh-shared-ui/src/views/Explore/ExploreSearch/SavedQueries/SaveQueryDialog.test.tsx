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
import { act, render, screen, waitFor } from '../../../../test-utils';
import { SavedQueriesProvider } from '../../providers';
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
    roles: ['Administrator'],
};

const testPermissions = {
    query_id: 3,
    public: true,
    shared_to_user_ids: [],
};

const handlers = [
    rest.get('/api/v2/bloodhound-users-minimal', (req, res, ctx) => {
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
    rest.get(`/api/v2/saved-queries/${undefined}/permissions`, async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: testPermissions,
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
    rest.get('/api/v2/features', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        id: 2,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'enable_saml_sso',
                        name: 'SAML Single Sign-On Support',
                        description:
                            'Enables SSO authentication flows and administration panels to third party SAML identity providers.',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 3,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'scope_collection_by_ou',
                        name: 'Enable SharpHound OU Scoped Collections',
                        description: 'Enables scoping SharpHound collections to specific lists of OUs.',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 4,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'azure_support',
                        name: 'Enable Azure Support',
                        description: 'Enables Azure support.',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 5,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'entity_panel_cache',
                        name: 'Enable application level caching',
                        description: 'Enables the use of application level caching for entity panel queries',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 6,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'adcs',
                        name: 'Enable collection and processing of Active Directory Certificate Services Data',
                        description:
                            'Enables the ability to collect, analyze, and explore Active Directory Certificate Services data and previews new attack paths.',
                        enabled: false,
                        user_updatable: false,
                    },
                    {
                        id: 7,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'dark_mode',
                        name: 'Dark Mode',
                        description: 'Allows users to enable or disable dark mode via a toggle in the settings menu',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 8,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'pg_migration_dual_ingest',
                        name: 'PostgreSQL Migration Dual Ingest',
                        description: 'Enables dual ingest pathing for both Neo4j and PostgreSQL.',
                        enabled: false,
                        user_updatable: false,
                    },
                    {
                        id: 9,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'clear_graph_data',
                        name: 'Clear Graph Data',
                        description: 'Enables the ability to delete all nodes and edges from the graph database.',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 10,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'risk_exposure_new_calculation',
                        name: 'Use new tier zero risk exposure calculation',
                        description: 'Enables the use of new tier zero risk exposure metatree metrics.',
                        enabled: false,
                        user_updatable: false,
                    },
                    {
                        id: 11,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'fedramp_eula',
                        name: 'FedRAMP EULA',
                        description: 'Enables showing the FedRAMP EULA on every login. (Enterprise only)',
                        enabled: false,
                        user_updatable: false,
                    },
                    {
                        id: 12,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'auto_tag_t0_parent_objects',
                        name: 'Automatically add parent OUs and containers of Tier Zero AD objects to Tier Zero',
                        description:
                            'Parent OUs and containers of Tier Zero AD objects are automatically added to Tier Zero during analysis. Containers are only added if they have a Tier Zero child object with ACL inheritance enabled.',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 18,
                        created_at: '2025-08-19T18:29:42.870942Z',
                        updated_at: '2025-08-19T18:29:42.870942Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'explore_table_view',
                        name: 'Explore Table View',
                        description:
                            'Adds a layout option to the Explore page that will display all nodes in a table view. It also will automatically display the table when a cypher query returned only nodes.',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 20,
                        created_at: '2025-08-19T18:29:42.884227Z',
                        updated_at: '2025-08-19T18:29:42.884227Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'targeted_access_control',
                        name: 'Targeted Access Control',
                        description: 'Enable power users and admins to set targeted access controls on users',
                        enabled: false,
                        user_updatable: false,
                    },
                    {
                        id: 13,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'oidc_support',
                        name: 'OIDC Support',
                        description: 'Enables OpenID Connect authentication support for SSO Authentication.',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 16,
                        created_at: '2025-08-19T18:29:42.857614Z',
                        updated_at: '2025-08-20T18:13:47.331002Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'tier_management_engine',
                        name: 'Tier Management Engine',
                        description: 'Updates the managed assets selector engine and the asset management page.',
                        enabled: true,
                        user_updatable: true,
                    },
                    {
                        id: 1,
                        created_at: '2025-08-19T18:29:42.724243Z',
                        updated_at: '2025-08-19T18:29:42.724243Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'butterfly_analysis',
                        name: 'Enhanced Asset Inbound-Outbound Exposure Analysis',
                        description:
                            'Enables more extensive analysis of attack path findings that allows BloodHound to help the user prioritize remediation of the most exposed assets.',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 17,
                        created_at: '2025-08-19T18:29:42.857614Z',
                        updated_at: '2025-08-19T18:29:42.857614Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'ntlm_post_processing',
                        name: 'NTLM Post Processing Support',
                        description:
                            'Enable the post processing of NTLM relay attack paths, this will enable the creation of CoerceAndRelayNTLMTo[LDAP, LDAPS, ADCS, SMB] edges.',
                        enabled: true,
                        user_updatable: false,
                    },
                    {
                        id: 15,
                        created_at: '2025-08-19T18:29:42.857614Z',
                        updated_at: '2025-08-19T18:29:42.857614Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        key: 'back_button_support',
                        name: 'Back Button Support',
                        description:
                            'Enable users to quickly navigate between views in a wider range of scenarios by utilizing the browser navigation buttons. Currently for BloodHound Community Edition users only.',
                        enabled: true,
                        user_updatable: false,
                    },
                ],
            })
        );
    }),
    rest.get('/api/v2/saved-queries', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        user_id: '65d70a82-5c54-48df-b172-a6d973cae737',
                        name: 'Custom Az Query',
                        query: 'MATCH p = (:User)-[:SyncedToEntraUser]->(:AZUser)-[:AZMemberOf]->(:AZGroup)\nRETURN p\nLIMIT 900',
                        description: '',
                        id: 34,
                        created_at: '2025-08-21T22:11:01.221359Z',
                        updated_at: '2025-08-22T15:21:27.530818Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        scope: 'owned',
                    },
                    {
                        user_id: '65d70a82-5c54-48df-b172-a6d973cae737',
                        name: 'Uploaded Query 1',
                        query: "MATCH p = (t:Group)<-[:MemberOf*1..]-(a) WHERE (a:User or a:Computer) and t.objectid ENDS WITH '-512' RETURN p LIMIT 101",
                        description: 'Uploaded Query 1 Desc',
                        id: 1,
                        created_at: '2025-08-19T18:32:09.287148Z',
                        updated_at: '2025-08-20T03:34:44.522163Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        scope: 'owned',
                    },
                    {
                        user_id: '65d70a82-5c54-48df-b172-a6d973cae737',
                        name: 'Uploaded Query 2 - update',
                        query: 'MATCH p = (:Domain)-[:SameForestTrust|CrossForestTrust]->(:Domain)\nRETURN p\nLIMIT 1022',
                        description: 'Uploaded Query 2 ',
                        id: 2,
                        created_at: '2025-08-19T18:32:09.287148Z',
                        updated_at: '2025-08-20T21:41:19.575045Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        scope: 'owned',
                    },
                    {
                        user_id: '65d70a82-5c54-48df-b172-a6d973cae737',
                        name: 'Uploaded Query 3',
                        query: "MATCH p = (t:Group)<-[:MemberOf*1..]-(a) WHERE (a:User or a:Computer) and t.objectid ENDS WITH '-512' RETURN p LIMIT 103",
                        description: 'Uploaded Query 3 Desc',
                        id: 3,
                        created_at: '2025-08-19T18:32:09.287148Z',
                        updated_at: '2025-08-19T18:32:09.287148Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        scope: 'owned',
                    },
                    {
                        user_id: '65d70a82-5c54-48df-b172-a6d973cae737',
                        name: 'Uploaded Query 4',
                        query: "MATCH p = (t:Group)<-[:MemberOf*1..]-(a) WHERE (a:User or a:Computer) and t.objectid ENDS WITH '-512' RETURN p LIMIT 104",
                        description: 'Uploaded Query 4 Desc',
                        id: 4,
                        created_at: '2025-08-19T18:32:09.307504Z',
                        updated_at: '2025-08-22T15:27:49.018151Z',
                        deleted_at: {
                            Time: '0001-01-01T00:00:00Z',
                            Valid: false,
                        },
                        scope: 'owned',
                    },
                ],
            })
        );
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('SaveQueryDialog', () => {
    const testOnSave = vi.fn();
    const testOnClose = vi.fn();
    const testOnUpdate = vi.fn();
    const testSetSharedIds = vi.fn();
    const testSetIsPublic = vi.fn();

    const testError = undefined;
    const testCypherSearchState = {
        cypherQuery: '',
        performSearch: vi.fn(),
        setCypherQuery: vi.fn(),
    };
    const SaveQueryDialogWithProvider = () => (
        <SavedQueriesProvider>
            <SaveQueryDialog
                open
                onSave={testOnSave}
                onClose={testOnClose}
                error={testError}
                cypherSearchState={testCypherSearchState}
                sharedIds={[]}
                isPublic={false}
                saveAction={undefined}
                saveUpdatePending={false}
                onUpdate={testOnUpdate}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        </SavedQueriesProvider>
    );

    it('should render a SaveQueryDialog', () => {
        render(<SaveQueryDialogWithProvider />);

        expect(screen.getByText(/save query/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/query name/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument();
    });

    it('should disable save button while input is empty', async () => {
        const user = userEvent.setup();
        const testQueryName = 'query name';

        render(<SaveQueryDialogWithProvider />);

        expect(screen.getByRole('button', { name: /save/i })).toBeDisabled();

        await user.type(screen.getByLabelText(/query name/i), testQueryName);

        expect(screen.getByRole('button', { name: /save/i })).not.toBeDisabled();
    });

    it('should call onClose when cancel button is clicked', async () => {
        const user = userEvent.setup();

        render(<SaveQueryDialogWithProvider />);

        await user.click(screen.getByRole('button', { name: /cancel/i }));

        expect(testOnClose).toHaveBeenCalled();
    });

    it('should call onSave when save button is clicked', async () => {
        const user = userEvent.setup();

        const testQueryName = 'query name';

        render(<SaveQueryDialogWithProvider />);

        await user.type(screen.getByLabelText(/query name/i), testQueryName);
        await user.click(screen.getByRole('button', { name: /save/i }));

        expect(testOnSave).toHaveBeenCalled();
    });

    it('should display an error when error prop is truthy', () => {
        const testError = true;
        render(
            <SavedQueriesProvider>
                <SaveQueryDialog
                    open
                    onSave={testOnSave}
                    onClose={testOnClose}
                    error={testError}
                    cypherSearchState={testCypherSearchState}
                    sharedIds={[]}
                    isPublic={false}
                    saveAction={undefined}
                    saveUpdatePending={false}
                    onUpdate={testOnUpdate}
                    setSharedIds={testSetSharedIds}
                    setIsPublic={testSetIsPublic}
                />
            </SavedQueriesProvider>
        );

        expect(
            screen.getByText(/an error occurred while attempting to save this query. please try again./i)
        ).toBeInTheDocument();
    });

    it('should render an SavedQueryPermissions component', async () => {
        await act(async () => render(<SaveQueryDialogWithProvider />));

        // Table Header Rendered
        const nestedElement = await waitFor(() => screen.getByText(/Manage Shared Queries/i));
        expect(nestedElement).toBeInTheDocument();
        screen.debug(nestedElement);
        const testTable = screen.getByRole('table');
        expect(testTable).toBeInTheDocument();
    });
});
