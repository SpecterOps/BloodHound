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

import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from 'src/test-utils';
import {
    getRACFUserCanSubmitAsQuery,
    getRACFUserClassAuthoritiesQuery,
    getRACFUserSubmittedAsByQuery,
    RACF_USER_CAN_SUBMIT_AS_SECTION,
    RACF_USER_CLASS_AUTHORITIES_SECTION,
    RACF_USER_INBOUND_RELATIONSHIPS_SECTION,
    RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION,
    RACF_USER_SUBMITTED_AS_BY_SECTION,
} from './groupMembers';
import { RACFUserInboundRelationships, RACFUserOutboundRelationships } from './RACFUserRelationships';

const graphResponse = (id: number, objectId: string, kind: string) => ({
    data: {
        nodes: {
            [id]: {
                objectId,
                label: objectId,
                kind,
                properties: { name: objectId, objectid: objectId },
            },
        },
        edges: [],
    },
});

const server = setupServer(
    rest.get('/api/v2/custom-nodes', (_request, response, context) => {
        return response(context.json({ data: [] }));
    }),
    rest.post('/api/v2/graphs/cypher', async (request, response, context) => {
        const body = await request.json();

        if (body.query.includes('ID(user) = 51')) {
            return response(context.status(404));
        }

        if (!body.query.includes('ID(user) = 50')) {
            return response(context.status(400));
        }

        if (body.query.includes('(user)-[:RACFSurrogateFor')) {
            return response(context.json(graphResponse(400, 'BATCHADM', 'RACFUser')));
        }

        if (body.query.includes('(principal)-[:RACFSurrogateFor')) {
            return response(context.json(graphResponse(401, 'APPDEV', 'RACFGroup')));
        }

        if (body.query.includes('RACFClassAuth')) {
            return response(context.json(graphResponse(402, 'FACILITY', 'RACFClass')));
        }

        return response(context.status(400));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('RACFUserRelationships', () => {
    it('groups outgoing SURROGAT and CLAUTH relationships under Outbound Relationships', async () => {
        const user = userEvent.setup();

        render(<RACFUserOutboundRelationships id='50' label={RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION} />);

        expect(await screen.findByText('2')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: new RegExp(RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION) }));
        await user.click(screen.getByRole('button', { name: new RegExp(RACF_USER_CAN_SUBMIT_AS_SECTION) }));

        expect(await screen.findByText('BATCHADM')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: new RegExp(RACF_USER_CLASS_AUTHORITIES_SECTION) }));

        expect(await screen.findByText('FACILITY')).toBeInTheDocument();
    });

    it('shows incoming SURROGAT principals under Inbound Relationships', async () => {
        const user = userEvent.setup();

        render(<RACFUserInboundRelationships id='50' label={RACF_USER_INBOUND_RELATIONSHIPS_SECTION} />);

        expect(await screen.findByText('1')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: new RegExp(RACF_USER_INBOUND_RELATIONSHIPS_SECTION) }));
        await user.click(screen.getByRole('button', { name: new RegExp(RACF_USER_SUBMITTED_AS_BY_SECTION) }));

        expect(await screen.findByText('APPDEV')).toBeInTheDocument();
    });

    it('renders an empty relationship group as a disabled zero-count section', async () => {
        render(<RACFUserOutboundRelationships id='51' label={RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION} />);

        expect(await screen.findByText('0')).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        expect(screen.getByRole('button')).toHaveAttribute('aria-disabled', 'true');
    });
});

describe('RACF user relationship queries', () => {
    it('uses the correct direct edge direction for each relationship', () => {
        expect(getRACFUserCanSubmitAsQuery('50')).toContain(
            '(user)-[:RACFSurrogateFor|racf_RACFSurrogateFor]->(target)'
        );
        expect(getRACFUserSubmittedAsByQuery('50')).toContain(
            '(principal)-[:RACFSurrogateFor|racf_RACFSurrogateFor]->(user)'
        );
        expect(getRACFUserClassAuthoritiesQuery('50')).toContain('(user)-[:RACFClassAuth|racf_RACFClassAuth]->(class)');
    });

    it('keeps direct relationship lists non-recursive', () => {
        expect(getRACFUserCanSubmitAsQuery('50')).not.toContain('*');
        expect(getRACFUserSubmittedAsByQuery('50')).not.toContain('*');
        expect(getRACFUserClassAuthoritiesQuery('50')).not.toContain('*');
    });

    it('rejects malformed graph database IDs', () => {
        expect(() => getRACFUserCanSubmitAsQuery('50 OR true')).toThrow('RACF user database ID must be an integer');
        expect(() => getRACFUserSubmittedAsByQuery('50 OR true')).toThrow('RACF user database ID must be an integer');
        expect(() => getRACFUserClassAuthoritiesQuery('50 OR true')).toThrow(
            'RACF user database ID must be an integer'
        );
    });
});
