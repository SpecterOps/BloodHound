// Copyright 2026 Specter Ops, Inc.
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

import { NodeDetails } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { ActiveDirectoryNodeKind } from '../../graphSchema';
import { mockSourceKindsHandler } from '../../mocks';
import { fireEvent, render, screen, waitFor } from '../../test-utils';
import { ObjectInfoPanelContextProvider } from '../../views';
import EntityObjectInformation from './EntityObjectInformation';

const server = setupServer(
    rest.get('/api/v2/features', (_req, res, ctx) => {
        return res(ctx.json({ data: [] }));
    }),
    rest.get('/api/v2/asset-group-tags', (_req, res, ctx) => {
        return res(ctx.json({ data: { tags: [] } }));
    }),
    rest.get('/api/v2/custom-nodes', (_req, res, ctx) => {
        return res(ctx.json({ data: [] }));
    }),
    mockSourceKindsHandler()
);

const EntityObjectInformationWithProvider = ({ selectedNode }: { selectedNode: NodeDetails }) => (
    <ObjectInfoPanelContextProvider>
        <EntityObjectInformation selectedNode={selectedNode} />
    </ObjectInfoPanelContextProvider>
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityObjectInformation', () => {
    it('renders the Object Information section and node properties', async () => {
        const selectedNode: NodeDetails = {
            node_id: 1,
            kinds: [{ name: ActiveDirectoryNodeKind.User, node_kind_id: 1 }],
            properties: { objectid: 'test-object-id', description: 'a test description' },
        };

        render(<EntityObjectInformationWithProvider selectedNode={selectedNode} />);

        expect(await screen.findByText('Object Information')).toBeInTheDocument();
        expect(screen.getByText('test-object-id')).toBeInTheDocument();
        expect(screen.getByText('a test description')).toBeInTheDocument();
    });

    it('renders and navigates to the Computer linked from a SiteServer through ServerIs', async () => {
        const computerObjectID = 'computer-object-id';
        const computerName = 'COMPUTER.TEST.LOCAL';
        const siteServerObjectID = 'site-server-object-id';

        server.use(
            rest.get(`/api/v2/siteservers/${siteServerObjectID}`, (_req, res, ctx) =>
                res(
                    ctx.json({
                        data: {
                            kinds: [ActiveDirectoryNodeKind.SiteServer],
                            props: {
                                objectid: siteServerObjectID,
                                serverreferencecomputer: computerObjectID,
                                serverreferencecomputername: computerName,
                            },
                        },
                    })
                )
            )
        );

        const selectedNode: NodeDetails = {
            node_id: 3,
            kinds: [{ name: ActiveDirectoryNodeKind.SiteServer, node_kind_id: 1 }],
            properties: { objectid: siteServerObjectID },
        };

        render(<EntityObjectInformationWithProvider selectedNode={selectedNode} />);

        expect(await screen.findByText('Referenced Computer:')).toBeInTheDocument();
        fireEvent.click(screen.getByText(computerName));

        await waitFor(() => {
            expect(window.location.search).toContain(`primarySearch=${computerObjectID}`);
            expect(window.location.search).toContain('searchType=node');
        });
    });

    it('renders the SiteServer linked from a Computer through ServerIs', async () => {
        const computerObjectID = 'computer-object-id';
        const siteServerName = 'SITE-SERVER.TEST.LOCAL';
        const siteServerObjectID = 'site-server-object-id';

        server.use(
            rest.get(`/api/v2/computers/${computerObjectID}`, (_req, res, ctx) =>
                res(
                    ctx.json({
                        data: {
                            kinds: [ActiveDirectoryNodeKind.Computer],
                            props: {
                                objectid: computerObjectID,
                                siteservernode: siteServerObjectID,
                                siteservernodename: siteServerName,
                            },
                        },
                    })
                )
            )
        );

        const selectedNode: NodeDetails = {
            node_id: 4,
            kinds: [{ name: ActiveDirectoryNodeKind.Computer, node_kind_id: 1 }],
            properties: { objectid: computerObjectID },
        };

        render(<EntityObjectInformationWithProvider selectedNode={selectedNode} />);

        expect(await screen.findByText('Site Server:')).toBeInTheDocument();
        expect(screen.getByText(siteServerName)).toBeInTheDocument();
    });

    it('does not render a linked node field when no ServerIs relationship exists', async () => {
        const requestSpy = vi.fn();
        const siteServerObjectID = 'unlinked-site-server-object-id';

        server.use(
            rest.get(`/api/v2/siteservers/${siteServerObjectID}`, (_req, res, ctx) => {
                requestSpy();
                return res(
                    ctx.json({
                        data: {
                            kinds: [ActiveDirectoryNodeKind.SiteServer],
                            props: { objectid: siteServerObjectID },
                        },
                    })
                );
            })
        );

        const selectedNode: NodeDetails = {
            node_id: 5,
            kinds: [{ name: ActiveDirectoryNodeKind.SiteServer, node_kind_id: 1 }],
            properties: { objectid: siteServerObjectID },
        };

        render(<EntityObjectInformationWithProvider selectedNode={selectedNode} />);

        await waitFor(() => expect(requestSpy).toHaveBeenCalledTimes(1));
        expect(screen.queryByText('Referenced Computer:')).not.toBeInTheDocument();
    });

    it('does not throw a React useRef error when a property is named "ref"', async () => {
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);

        try {
            const selectedNode: NodeDetails = {
                node_id: 2,
                kinds: [{ name: ActiveDirectoryNodeKind.User, node_kind_id: 1 }],
                properties: { objectid: 'ref-object-id', ref: 'a-property-named-ref' },
            };

            expect(() => render(<EntityObjectInformationWithProvider selectedNode={selectedNode} />)).not.toThrow();

            // The value of the "ref" property should render as a normal field value.
            expect(await screen.findByText('a-property-named-ref')).toBeInTheDocument();

            // React logs ref-related warnings/errors (e.g. useRef, forwardRef, "Function
            // components cannot be given refs") through console.error. Ensure none occurred.
            const refRelatedErrors = consoleErrorSpy.mock.calls.filter((args) =>
                args.some((arg) => typeof arg === 'string' && /useRef|forwardRef|given refs/i.test(arg))
            );
            expect(refRelatedErrors).toHaveLength(0);
        } finally {
            consoleErrorSpy.mockRestore();
        }
    });
});
