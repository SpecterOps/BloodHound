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
    getRACFGroupCanSubmitAsQuery,
    RACF_GROUP_CAN_SUBMIT_AS_SECTION,
    RACF_GROUP_OUTBOUND_RELATIONSHIPS_SECTION,
} from './groupMembers';
import { RACFGroupOutboundRelationships } from './RACFGroupRelationships';

const server = setupServer(
    rest.get('/api/v2/custom-nodes', (_request, response, context) => {
        return response(context.json({ data: [] }));
    }),
    rest.post('/api/v2/graphs/cypher', async (request, response, context) => {
        const body = await request.json();

        if (body.query.includes('ID(group) = 71')) {
            return response(context.status(404));
        }

        if (!body.query.includes('ID(group) = 70') || !body.query.includes('RACFSurrogateFor')) {
            return response(context.status(400));
        }

        return response(
            context.json({
                data: {
                    nodes: {
                        600: {
                            objectId: 'BATCHADM',
                            label: 'BATCHADM',
                            kind: 'RACFUser',
                            properties: { name: 'BATCHADM', objectid: 'BATCHADM' },
                        },
                    },
                    edges: [],
                },
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('RACFGroupOutboundRelationships', () => {
    it('shows users that the group can submit as', async () => {
        const user = userEvent.setup();

        render(<RACFGroupOutboundRelationships id='70' label={RACF_GROUP_OUTBOUND_RELATIONSHIPS_SECTION} />);

        expect(await screen.findByText('1')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: new RegExp(RACF_GROUP_OUTBOUND_RELATIONSHIPS_SECTION) }));
        await user.click(screen.getByRole('button', { name: new RegExp(RACF_GROUP_CAN_SUBMIT_AS_SECTION) }));

        expect(await screen.findByText('BATCHADM')).toBeInTheDocument();
    });

    it('shows a disabled zero-count section when the group has no SURROGAT relationships', async () => {
        render(<RACFGroupOutboundRelationships id='71' label={RACF_GROUP_OUTBOUND_RELATIONSHIPS_SECTION} />);

        expect(await screen.findByText('0')).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        expect(screen.getByRole('button')).toHaveAttribute('aria-disabled', 'true');
    });
});

describe('RACF group SURROGAT query', () => {
    it('uses a direct outgoing relationship to the target user', () => {
        const query = getRACFGroupCanSubmitAsQuery('70');

        expect(query).toContain('(group)-[:RACFSurrogateFor|racf_RACFSurrogateFor]->(target)');
        expect(query).not.toContain('*');
    });

    it('rejects malformed graph database IDs', () => {
        expect(() => getRACFGroupCanSubmitAsQuery('70 OR true')).toThrow('RACF group database ID must be an integer');
    });
});
