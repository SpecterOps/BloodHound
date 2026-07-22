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
import { RelationshipDetails } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { INHERITANCE_DROPDOWN_DESCRIPTION } from '../../../components/HelpTexts/shared/ACLInheritance';
import {
    ActiveDirectoryKindProperties,
    ActiveDirectoryRelationshipKind,
    CommonKindProperties,
} from '../../../graphSchema';
import { mockSourceKindsHandler } from '../../../mocks';
import { render, screen, waitFor } from '../../../test-utils';
import { ObjectInfoPanelContextProvider } from '../providers';
import EdgeInfoContent from './EdgeInfoContent';

const server = setupServer(
    rest.get(`/api/v2/relationships/1`, (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    relationship_id: 1,
                    kind: { relationship_kind_id: 1, name: 'CustomEdge' },
                    source_node_id: 1,
                    target_node_id: 2,
                    properties: {
                        isacl: false,
                        is_traversable: true,
                        lastSeen: '2023-09-07T11:10:33.664596893Z',
                    },
                },
            })
        );
    }),
    rest.get(`/api/v2/nodes/1`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    node_id: 1,
                    kinds: [{ node_kind_id: 1, name: 'User' }],
                    properties: {
                        objectid: 'source-node-1',
                        name: 'Source User',
                        lastSeen: '2023-09-07T11:10:33.664596893Z',
                    },
                },
            })
        );
    }),
    rest.get(`/api/v2/nodes/2`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    node_id: 2,
                    kinds: [{ node_kind_id: 2, name: 'User' }],
                    properties: {
                        objectid: 'target-node-2',
                        name: 'Target User',
                        lastSeen: '2023-09-07T11:10:33.664596893Z',
                    },
                },
            })
        );
    }),
    rest.get(`/api/v2/nodes/3`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    node_id: 3,
                    kinds: [{ node_kind_id: 3, name: 'Computer' }],
                    properties: {
                        objectid: 'target-node-3',
                        name: 'Target Computer With LAPS',
                        haslaps: true,
                        lastSeen: '2023-09-07T11:10:33.664596893Z',
                    },
                },
            })
        );
    }),
    rest.get(`/api/v2/nodes/4`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    node_id: 4,
                    kinds: [{ node_kind_id: 4, name: 'Computer' }],
                    properties: {
                        objectid: 'target-node-4',
                        name: 'Target Computer Without LAPS',
                        lastSeen: '2023-09-07T11:10:33.664596893Z',
                    },
                },
            })
        );
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    }),
    rest.get('/api/v2/graphs/acl-inheritance', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    nodes: {},
                    edges: [],
                },
            })
        );
    }),
    mockSourceKindsHandler()
);
const selectedEdge: RelationshipDetails = {
    relationship_id: 1,
    kind: { name: 'CustomEdge', relationship_kind_id: 1 },
    properties: {
        lastSeen: '',
        is_traversable: false,
        [ActiveDirectoryKindProperties.IsACL]: false,
        [CommonKindProperties.LastSeen]: '2023-09-07T11:10:33.664596893Z',
    },
    source_node_id: 1,
    target_node_id: 2,
};

const selectedEdgeHasLapsEnabled: RelationshipDetails = {
    relationship_id: 2,
    kind: { name: ActiveDirectoryRelationshipKind.GenericAll, relationship_kind_id: 2 },
    properties: {
        lastSeen: '',
        is_traversable: false,
        [ActiveDirectoryKindProperties.IsACL]: false,
        [CommonKindProperties.LastSeen]: '2023-09-07T11:10:33.664596893Z',
    },
    source_node_id: 1,
    target_node_id: 3,
};

const selectedEdgeHasLapsDisabled: RelationshipDetails = {
    relationship_id: 3,
    kind: { name: ActiveDirectoryRelationshipKind.GenericAll, relationship_kind_id: 2 },
    properties: {
        lastSeen: '',
        is_traversable: false,
        [ActiveDirectoryKindProperties.IsACL]: false,
        [CommonKindProperties.LastSeen]: '2023-09-07T11:10:33.664596893Z',
    },
    source_node_id: 1,
    target_node_id: 4,
};

const selectedEdgeADCSESC4: RelationshipDetails = {
    ...selectedEdge,
    kind: { name: ActiveDirectoryRelationshipKind.ADCSESC4, relationship_kind_id: 4 },
};

const selectedEdgeACLInheritance: RelationshipDetails = {
    relationship_id: 2,
    kind: { name: ActiveDirectoryRelationshipKind.GenericAll, relationship_kind_id: 2 },
    properties: {
        lastSeen: '',
        is_traversable: false,
        [ActiveDirectoryKindProperties.IsACL]: true,
        [CommonKindProperties.LastSeen]: '2023-09-07T11:10:33.664596893Z',
        [CommonKindProperties.IsInherited]: true,
        [ActiveDirectoryKindProperties.InheritanceHash]: 'test_hash',
    },
    source_node_id: 1,
    target_node_id: 4,
};

const windowsAbuseHasLapsText = (sourceName: string, targetName: string) => {
    return `The GenericAll permission grants ${sourceName} the ability to obtain the LAPS (RID 500 administrator) password of ${targetName}.`;
};

// Node names are sourced from the MSW mock handlers for /api/v2/nodes/:id
const hasLapsEnabledTestText = windowsAbuseHasLapsText('Source User', 'Target Computer With LAPS');
const hasLapsDisabledTestText = windowsAbuseHasLapsText('Source User', 'Target Computer Without LAPS');

const EdgeInfoContentWithProvider = ({ selectedEdge }: { selectedEdge: RelationshipDetails }) => (
    <ObjectInfoPanelContextProvider>
        <EdgeInfoContent selectedEdge={selectedEdge!} />
    </ObjectInfoPanelContextProvider>
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EdgeInfoContent', () => {
    test('Trying to view the edge info does not crash the app when selecting an unrecognized edge', async () => {
        render(<EdgeInfoContentWithProvider selectedEdge={selectedEdge} />);

        expect(await screen.findByText(/Source Node:/)).toBeInTheDocument();

        //The general object information is still available even though there is no contextual information available for the edge
        expect(screen.getByText(/Source Node:/)).toBeInTheDocument();
        expect(screen.getByText(/Target Node:/)).toBeInTheDocument();
        expect(screen.getByText(/Is ACL:/)).toBeInTheDocument();
        expect(screen.getByText(/Last Seen by BloodHound:/)).toBeInTheDocument();

        //The whole app does not crash and require a refresh
        expect(
            screen.queryByText('An unexpected error has occurred. Please refresh the page and try again.')
        ).not.toBeInTheDocument();
    });
    test('Selecting an edge with a Computer target node that haslaps is enabled shows correct Windows Abuse text', async () => {
        render(<EdgeInfoContentWithProvider selectedEdge={selectedEdgeHasLapsEnabled} />);

        const user = userEvent.setup();
        const windowAbuseAccordion = await screen.findByText('Windows Abuse');
        await user.click(windowAbuseAccordion);

        expect(screen.getByText(hasLapsEnabledTestText, { exact: false })).toBeInTheDocument();
    });
    test('Selecting an edge with a Computer target node that does not have haslaps enabled shows correct Windows Abuse text', async () => {
        render(<EdgeInfoContentWithProvider selectedEdge={selectedEdgeHasLapsDisabled} />);

        const user = userEvent.setup();
        const windowAbuseAccordion = await screen.findByText('Windows Abuse');
        await user.click(windowAbuseAccordion);

        expect(screen.queryByText(hasLapsDisabledTestText, { exact: false })).not.toBeInTheDocument();
    });
    test('Selecting an edge that meets ACL inheritance criteria shows the "ACE Inherited From" dropdown', async () => {
        render(<EdgeInfoContentWithProvider selectedEdge={selectedEdgeACLInheritance} />);

        const user = userEvent.setup();
        const inheritanceAccordion = await screen.findByText('ACE Inherited From');
        await user.click(inheritanceAccordion);

        expect(screen.queryByText(INHERITANCE_DROPDOWN_DESCRIPTION)).toBeInTheDocument();
    });
    describe('EdgeInfoContent support for Deep Linking', () => {
        const test_id = selectedEdgeADCSESC4.relationship_id;
        const setup = () => {
            const screen = render(<EdgeInfoContentWithProvider selectedEdge={selectedEdgeADCSESC4} />, {
                route: `?selectedItem=${test_id}`,
            });
            const user = userEvent.setup();

            server.use(
                rest.get('/api/v2/graphs/edge-composition', (req, res, ctx) => {
                    return res(
                        ctx.json({
                            data: [],
                        })
                    );
                })
            );

            return { screen, user };
        };

        it('calls setExploreParams with searchType and relationshipQueryItemId when selecting a composition accordion', async () => {
            const { user, screen } = setup();

            const compositionAccordion = await screen.findByText('Composition');
            await user.click(compositionAccordion);

            await waitFor(() => {
                expect(window.location.search).toContain('searchType=composition');
            });
            expect(window.location.search).toContain(`relationshipQueryItemId=rel_${test_id}`);
        });
        it('calls setExploreParams with only the expandedSection label when selecting any accordion that is not composition', async () => {
            const { user, screen } = setup();

            const generalAccordion = await screen.findByText('General');
            await user.click(generalAccordion);

            await waitFor(() => expect(window.location.search).toContain('expandedPanelSections=general'));
            expect(window.location.search).not.toContain('searchType');
            expect(window.location.search).not.toContain(`relationshipQueryItemId=${test_id}`);
        });
    });

    describe('EdgeInfoContent support for hidden edges', () => {
        const setup = () => {
            // isHidden is derived from selectedItem URL param containing 'HIDDEN'
            const screen = render(<EdgeInfoContentWithProvider selectedEdge={selectedEdge} />, {
                route: '?selectedItem=HIDDEN',
            });
            const user = userEvent.setup();

            return { screen, user };
        };

        it('displays contact admin message when hidden edge is true', async () => {
            const { screen } = setup();

            expect(
                await screen.findByText(
                    "This edge's information is not disclosed. Please contact your admin in order to get access."
                )
            ).toBeInTheDocument();
        });
    });
});
