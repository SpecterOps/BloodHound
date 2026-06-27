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
import { RACFClassUsersWithCLAUTH } from './RACFClassRelationships';
import {
    getRACFClassUsersWithCLAUTHQuery,
    isRACFClassKind,
    RACF_CLASS_USERS_WITH_CLAUTH_SECTION,
} from './groupMembers';

const server = setupServer(
    rest.get('/api/v2/custom-nodes', (_request, response, context) => {
        return response(context.json({ data: [] }));
    }),
    rest.post('/api/v2/graphs/cypher', async (request, response, context) => {
        const body = await request.json();

        if (body.query.includes('ID(class) = 61')) {
            return response(context.status(404));
        }

        if (!body.query.includes('ID(class) = 60') || !body.query.includes('RACFClassAuth')) {
            return response(context.status(400));
        }

        return response(
            context.json({
                data: {
                    nodes: {
                        500: {
                            objectId: 'USR001',
                            label: 'USR001',
                            kind: 'RACFUser',
                            properties: { name: 'USR001', objectid: 'USR001' },
                        },
                        501: {
                            objectId: 'SECADMIN',
                            label: 'SECADMIN',
                            kind: 'RACFUser',
                            properties: { name: 'SECADMIN', objectid: 'SECADMIN' },
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

describe('RACFClassUsersWithCLAUTH', () => {
    it('shows users with direct CLAUTH to the selected class', async () => {
        const user = userEvent.setup();

        render(<RACFClassUsersWithCLAUTH id='60' label={RACF_CLASS_USERS_WITH_CLAUTH_SECTION} />);

        expect(await screen.findByText('2')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: new RegExp(RACF_CLASS_USERS_WITH_CLAUTH_SECTION) }));

        expect(await screen.findByText('USR001')).toBeInTheDocument();
        expect(await screen.findByText('SECADMIN')).toBeInTheDocument();
    });

    it('shows a disabled zero-count section when nobody has CLAUTH', async () => {
        render(<RACFClassUsersWithCLAUTH id='61' label={RACF_CLASS_USERS_WITH_CLAUTH_SECTION} />);

        expect(await screen.findByText('0')).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        expect(screen.getByRole('button')).toHaveAttribute('aria-disabled', 'true');
    });
});

describe('RACF class relationship helpers', () => {
    it.each(['RACFClass', 'racf_RACFClass', 'racfclass'])('recognizes the %s class kind', (kind) => {
        expect(isRACFClassKind(kind)).toBe(true);
    });

    it('queries incoming direct CLAUTH relationships', () => {
        const query = getRACFClassUsersWithCLAUTHQuery('60');

        expect(query).toContain('(user)-[:RACFClassAuth|racf_RACFClassAuth]->(class)');
        expect(query).not.toContain('*');
    });

    it('rejects malformed graph database IDs', () => {
        expect(() => getRACFClassUsersWithCLAUTHQuery('60 OR true')).toThrow(
            'RACF class database ID must be an integer'
        );
    });
});
