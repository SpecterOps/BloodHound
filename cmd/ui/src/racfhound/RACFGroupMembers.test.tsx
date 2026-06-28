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
import { RACFGroupMembers, RACFGroupSubgroups, RACFUserGroups } from './RACFGroupMembers';
import {
    getRACFGroupMembersQuery,
    getRACFGroupSubgroupsQuery,
    getRACFUserGroupsQuery,
    isRACFGroupKind,
    isRACFUserKind,
    RACF_GROUP_MEMBERS_SECTION,
    RACF_GROUP_SUBGROUPS_SECTION,
    RACF_USER_GROUPS_SECTION,
} from './groupMembers';

const server = setupServer(
    rest.get('/api/v2/custom-nodes', (_request, response, context) => {
        return response(context.json({ data: [] }));
    }),
    rest.post('/api/v2/graphs/cypher', async (request, response, context) => {
        const body = await request.json();

        if (body.query.includes('ID(group) = 43') || body.query.includes('ID(group) = 44')) {
            return response(context.status(404));
        }

        if (body.query.includes('ID(user) = 46')) {
            return response(context.status(404));
        }

        if (body.query.includes('ID(user) = 45') && body.query.includes('RACFMemberOf')) {
            return response(
                context.json({
                    data: {
                        nodes: {
                            300: {
                                objectId: 'APPDEV',
                                label: 'APPDEV',
                                kind: 'RACFGroup',
                                properties: { name: 'APPDEV', objectid: 'APPDEV' },
                            },
                        },
                        edges: [],
                    },
                })
            );
        }

        if (!body.query.includes('ID(group) = 42')) {
            return response(context.status(400));
        }

        if (body.query.includes('RACFHasSubgroup')) {
            return response(
                context.json({
                    data: {
                        nodes: {
                            200: {
                                objectId: 'NESTED1',
                                label: 'NESTED1',
                                kind: 'RACFGroup',
                                properties: { name: 'NESTED1', objectid: 'NESTED1' },
                            },
                        },
                        edges: [],
                    },
                })
            );
        }

        if (!body.query.includes('RACFMemberOf')) {
            return response(context.status(400));
        }

        return response(
            context.json({
                data: {
                    nodes: {
                        100: {
                            objectId: 'USR001',
                            label: 'USR001',
                            kind: 'RACFUser',
                            properties: { name: 'USR001', objectid: 'USR001' },
                        },
                        101: {
                            objectId: 'USR002',
                            label: 'USR002',
                            kind: 'RACFUser',
                            properties: { name: 'USR002', objectid: 'USR002' },
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

describe('RACFGroupMembers', () => {
    it('shows direct RACF group members', async () => {
        const user = userEvent.setup();

        render(<RACFGroupMembers id='42' label={RACF_GROUP_MEMBERS_SECTION} />);

        expect(await screen.findByText('2')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: new RegExp(RACF_GROUP_MEMBERS_SECTION) }));

        expect(await screen.findByText('USR001')).toBeInTheDocument();
        expect(await screen.findByText('USR002')).toBeInTheDocument();
    });

    it('shows direct subgroups in a separate section', async () => {
        const user = userEvent.setup();

        render(<RACFGroupSubgroups id='42' label={RACF_GROUP_SUBGROUPS_SECTION} />);

        expect(await screen.findByText('1')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: new RegExp(RACF_GROUP_SUBGROUPS_SECTION) }));

        expect(await screen.findByText('NESTED1')).toBeInTheDocument();
    });

    it.each([
        ['members', RACFGroupMembers, '43', RACF_GROUP_MEMBERS_SECTION],
        ['subgroups', RACFGroupSubgroups, '44', RACF_GROUP_SUBGROUPS_SECTION],
    ])('shows a zero count instead of an error when a group has no %s', async (_name, Component, id, label) => {
        render(<Component id={id} label={label} />);

        expect(await screen.findByText('0')).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        expect(screen.getByRole('button')).toHaveAttribute('aria-disabled', 'true');
    });
});

describe('RACFUserGroups', () => {
    it('shows groups directly connected to the RACF user', async () => {
        const user = userEvent.setup();

        render(<RACFUserGroups id='45' label={RACF_USER_GROUPS_SECTION} />);

        expect(await screen.findByText('1')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: new RegExp(RACF_USER_GROUPS_SECTION) }));

        expect(await screen.findByText('APPDEV')).toBeInTheDocument();
    });

    it('shows a disabled zero-count section when the user has no groups', async () => {
        render(<RACFUserGroups id='46' label={RACF_USER_GROUPS_SECTION} />);

        expect(await screen.findByText('0')).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        expect(screen.getByRole('button')).toHaveAttribute('aria-disabled', 'true');
    });
});

describe('RACF group member helpers', () => {
    it.each(['RACFGroup', 'racfgroup'])('recognizes the %s group kind', (kind) => {
        expect(isRACFGroupKind(kind)).toBe(true);
    });

    it.each(['RACFUser', 'racfuser'])('recognizes the %s user kind', (kind) => {
        expect(isRACFUserKind(kind)).toBe(true);
    });

    it('rejects a non-numeric graph database ID', () => {
        expect(() => getRACFGroupMembersQuery('42 OR true')).toThrow('RACF group database ID must be an integer');
        expect(() => getRACFGroupSubgroupsQuery('42 OR true')).toThrow('RACF group database ID must be an integer');
        expect(() => getRACFUserGroupsQuery('42 OR true')).toThrow('RACF user database ID must be an integer');
    });

    it('does not traverse subgroups when querying members', () => {
        expect(getRACFGroupMembersQuery('42')).not.toContain('RACFHasSubgroup');
        expect(getRACFGroupSubgroupsQuery('42')).not.toContain('*');
    });

    it('queries only groups directly connected to the user', () => {
        const query = getRACFUserGroupsQuery('45');

        expect(query).toContain('(user)-[:RACFMemberOf]->(group)');
        expect(query).not.toContain('*');
    });
});
