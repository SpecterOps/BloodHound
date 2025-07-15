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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { INHERITANCE_DROPDOWN_DESCRIPTION } from '../../../components/HelpTexts/shared/ACLInheritance';
import {
    ActiveDirectoryKindProperties,
    ActiveDirectoryNodeKind,
    ActiveDirectoryRelationshipKind,
    CommonKindProperties,
} from '../../../graphSchema';
import { SelectedEdge } from '../../../store';
import { render, screen, waitFor } from '../../../test-utils';
import { ObjectInfoPanelContextProvider } from '../providers';
import EdgeInfoContent from './EdgeInfoContent';

const server = setupServer(
    rest.post(`/api/v2/graphs/cypher`, (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    nodes: {},
                    edges: [
                        {
                            source: '1',
                            target: '2',
                            label: 'CustomEdge',
                            kind: 'CustomEdge',
                            lastSeen: '2023-09-07T11:10:33.664596893Z',
                            properties: {
                                lastseen: '2023-09-07T11:10:33.664596893Z',
                                isacl: false,
                            },
                        },
                    ],
                },
            })
        );
    }),
    rest.get(`/api/v2/users/:id`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    props: {
                        objectid: '2',
                    },
                },
            })
        );
    }),
    rest.get(`/api/v2/computers/testing-node-123`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    props: {
                        haslaps: true,
                        objectid: 'testing-node-123',
                    },
                },
            })
        );
    }),
    rest.get(`/api/v2/computers/testing-node-456`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    props: {
                        objectid: 'testing-node-456',
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
    })
);

const selectedEdge: SelectedEdge = {
    id: 'rel_1',
    name: 'CustomEdge',
    data: {
        [ActiveDirectoryKindProperties.IsACL]: false,
        [CommonKindProperties.LastSeen]: '2023-09-07T11:10:33.664596893Z',
    },
    sourceNode: {
        name: 'source node',
        id: '1',
        objectId: '1',
        type: ActiveDirectoryNodeKind.User,
    },
    targetNode: {
        name: 'target node',
        id: '2',
        objectId: '2',
        type: ActiveDirectoryNodeKind.User,
    },
};

const selectedEdgeHasLapsEnabled: SelectedEdge = {
    id: '2',
    name: ActiveDirectoryRelationshipKind.GenericAll,
    data: {
        [ActiveDirectoryKindProperties.IsACL]: false,
        [CommonKindProperties.LastSeen]: '2023-09-07T11:10:33.664596893Z',
    },
    sourceNode: {
        name: 'source node',
        id: '1',
        objectId: '1',
        type: ActiveDirectoryNodeKind.User,
    },
    targetNode: {
        name: 'target node',
        id: '3',
        objectId: 'testing-node-123',
        type: ActiveDirectoryNodeKind.Computer,
    },
};

const selectedEdgeHasLapsDisabled: SelectedEdge = {
    id: '3',
    name: ActiveDirectoryRelationshipKind.GenericAll,
    data: {
        [ActiveDirectoryKindProperties.IsACL]: false,
        [CommonKindProperties.LastSeen]: '2023-09-07T11:10:33.664596893Z',
    },
    sourceNode: {
        name: 'source node',
        id: '1',
        objectId: '1',
        type: ActiveDirectoryNodeKind.User,
    },
    targetNode: {
        name: 'target node',
        id: '4',
        objectId: 'testing-node-456',
        type: ActiveDirectoryNodeKind.Computer,
    },
};

const selectedEdgeADCSESC4: SelectedEdge = {
    ...selectedEdge,
    name: ActiveDirectoryRelationshipKind.ADCSESC4,
};

const selectedEdgeACLInheritance: SelectedEdge = {
    id: '3',
    name: ActiveDirectoryRelationshipKind.WriteOwner,
    data: {
        [ActiveDirectoryKindProperties.IsACL]: true,
        [CommonKindProperties.IsInherited]: true,
        [ActiveDirectoryKindProperties.InheritanceHash]: 'test_hash',
    },
    sourceNode: {
        name: 'source node',
        id: '1',
        objectId: '1',
        type: ActiveDirectoryNodeKind.Group,
    },
    targetNode: {
        name: 'target node',
        id: '4',
        objectId: 'testing-node-456',
        type: ActiveDirectoryNodeKind.Group,
    },
};

const windowsAbuseHasLapsText = (sourceName: string, targetName: string) => {
    return `The GenericAll permission grants ${sourceName} the ability to obtain the LAPS (RID 500 administrator) password of ${targetName}.`;
};

const hasLapsEnabledTestText = windowsAbuseHasLapsText(
    selectedEdgeHasLapsEnabled.sourceNode.name,
    selectedEdgeHasLapsEnabled.targetNode.name
);
const hasLapsDisabledTestText = windowsAbuseHasLapsText(
    selectedEdgeHasLapsDisabled.sourceNode.name,
    selectedEdgeHasLapsDisabled.targetNode.name
);

const EdgeInfoContentWithProvider = ({ selectedEdge }: { selectedEdge: SelectedEdge }) => (
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

        expect(await screen.findByText(/source node/)).toBeInTheDocument();

        //The text is broken up into different elements because of the span that bolds the custom edge so these assertions are broken up to match that
        //The assertions use regex to avoid having to match on the white space
        expect(screen.getByText(/The edge/)).toBeInTheDocument();
        expect(screen.getByText(selectedEdge.name)).toBeInTheDocument();
        expect(
            screen.getByText(/does not have any additional contextual information at this time./)
        ).toBeInTheDocument();

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
        const windowAbuseAccordion = screen.getByText('Windows Abuse');
        await user.click(windowAbuseAccordion);

        expect(screen.getByText(hasLapsEnabledTestText, { exact: false })).toBeInTheDocument();
    });
    test('Selecting an edge with a Computer target node that does not have haslaps enabled shows correct Windows Abuse text', async () => {
        render(<EdgeInfoContentWithProvider selectedEdge={selectedEdgeHasLapsDisabled} />);

        const user = userEvent.setup();
        const windowAbuseAccordion = screen.getByText('Windows Abuse');
        await user.click(windowAbuseAccordion);

        expect(screen.queryByText(hasLapsDisabledTestText, { exact: false })).not.toBeInTheDocument();
    });
    test('Selecting an edge that meets ACL inheritance criteria shows the "ACE Inherited From" dropdown', async () => {
        render(<EdgeInfoContentWithProvider selectedEdge={selectedEdgeACLInheritance} />);

        const user = userEvent.setup();
        const inheritanceAccordion = screen.getByText('ACE Inherited From');
        await user.click(inheritanceAccordion);

        expect(screen.queryByText(INHERITANCE_DROPDOWN_DESCRIPTION)).toBeInTheDocument();
    });
    describe('EdgeInfoContent support for Deep Linking', () => {
        const test_id = selectedEdgeADCSESC4.id;
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

            const compositionAccordion = screen.getByText('Composition');
            await user.click(compositionAccordion);

            await waitFor(() => {
                expect(window.location.search).toContain('searchType=composition');
            });
            expect(window.location.search).toContain(`relationshipQueryItemId=${test_id}`);
        });
        it('calls setExploreParams with only the expandedSection label when selecting any accordion that is not composition', async () => {
            const { user, screen } = setup();

            const generalAccordion = screen.getByText('General');
            await user.click(generalAccordion);

            await waitFor(() => expect(window.location.search).toContain('expandedPanelSections=general'));
            expect(window.location.search).not.toContain('searchType');
            expect(window.location.search).not.toContain(`relationshipQueryItemId=${test_id}`);
        });
    });
});
