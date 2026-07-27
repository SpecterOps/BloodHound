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

import {
    RACF_CLASS_USERS_WITH_CLAUTH_SECTION,
    RACF_GROUP_MEMBERS_SECTION,
    RACF_GROUP_OUTBOUND_RELATIONSHIPS_SECTION,
    RACF_GROUP_SUBGROUPS_SECTION,
    RACF_USER_GROUPS_SECTION,
    RACF_USER_INBOUND_RELATIONSHIPS_SECTION,
    RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION,
} from './groupMembers';
import { getRACFAdditionalTables } from './racfAdditionalTables';

const labelsOf = (kindName: string, id = '105') =>
    getRACFAdditionalTables({ kinds: [{ name: kindName }] }, id)?.map((table) => table.sectionProps.label);

describe('getRACFAdditionalTables', () => {
    it('returns the user relationship sections for a RACFUser node', () => {
        expect(labelsOf('RACFUser')).toEqual([
            RACF_USER_GROUPS_SECTION,
            RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION,
            RACF_USER_INBOUND_RELATIONSHIPS_SECTION,
        ]);
    });

    it('returns the group relationship sections for a RACFGroup node', () => {
        expect(labelsOf('RACFGroup')).toEqual([
            RACF_GROUP_MEMBERS_SECTION,
            RACF_GROUP_SUBGROUPS_SECTION,
            RACF_GROUP_OUTBOUND_RELATIONSHIPS_SECTION,
        ]);
    });

    it('returns the CLAUTH section for a RACFClass node', () => {
        expect(labelsOf('RACFClass')).toEqual([RACF_CLASS_USERS_WITH_CLAUTH_SECTION]);
    });

    // Guards the `kinds: NodeKindRef[]` shape: kinds are objects with a `name`, not bare strings.
    it('reads the kind from the .name of each NodeKindRef and matches even when RACF is not the first kind', () => {
        const tables = getRACFAdditionalTables({ kinds: [{ name: 'Base' }, { name: 'RACFGroup' }] }, '7');
        expect(tables?.map((table) => table.sectionProps.label)).toEqual([
            RACF_GROUP_MEMBERS_SECTION,
            RACF_GROUP_SUBGROUPS_SECTION,
            RACF_GROUP_OUTBOUND_RELATIONSHIPS_SECTION,
        ]);
    });

    it('passes the supplied database id through to every section', () => {
        const tables = getRACFAdditionalTables({ kinds: [{ name: 'RACFUser' }] }, '4242');
        expect(tables?.every((table) => table.sectionProps.id === '4242')).toBe(true);
    });

    it('returns undefined for non-RACF nodes', () => {
        expect(getRACFAdditionalTables({ kinds: [{ name: 'User' }] }, '1')).toBeUndefined();
    });

    it('returns undefined when the node has no kinds', () => {
        expect(getRACFAdditionalTables({}, '1')).toBeUndefined();
    });
});
