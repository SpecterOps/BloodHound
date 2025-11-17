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

import { faker } from '@faker-js/faker/locale/en';
import { ActiveDirectoryNodeKind } from 'bh-shared-ui';

export const createDatapipeStatus = (updated_at?: Date) => {
    return {
        start: '2022-03-23T07:20:50.52Z',
        end: '2022-03-23T07:20:50.52Z',
        data: {
            status: faker.helpers.arrayElement(['idle', 'ingesting', 'analyzing']),
            updated_at: updated_at || '2022-05-09T18:24:19.3856229Z',
        },
    };
};

export const createEntityResponse = () => {
    return {
        adminRights: 0,
        adminUsers: 54,
        constrainedPrivs: 0,
        constrainedUsers: 0,
        controllables: 0,
        controllers: 52,
        dcomRights: 0,
        dcomUsers: 0,
        gpos: 1,
        groupMembership: 1,
        props: {
            testArray: ['multi', 'values', 'for', 'testing'],
            testArrayAdjacent: ['multi', 'values'],
            extraProp: 'test',
            testArray2: ['multi2', 'values2', 'for2', 'testing2'],
            extraPropNum: 7357,
            testEmptyArray: [],
            domain: 'TESTLAB.LOCAL',
            domainsid: 'S-1-5-21-570004220-2248230615-4072641716',
            enabled: true,
            lastseen: '2022-09-20T19:21:25.3391482Z',
            name: '00001.TESTLAB.LOCAL',
            objectid: 'S-1-5-21-570004220-2248230615-4072641716-1000',
            operatingsystem: 'Windows Server 2008',
            pwdlastset: 1663701685,
        },
        psRemoteRights: 0,
        psRemoteUsers: 0,
        backupRights: 0,
        backupOperators: 0,
        rdpRights: 0,
        rdpUsers: 0,
        sessions: 0,
        sqlAdminUsers: 0,
    };
};

export const createMockSearchResult = (nodeType?: string) => {
    const name = faker.random.word();
    return {
        objectid: faker.datatype.uuid(),
        name: name,
        distinguishedname: name,
        type:
            nodeType ||
            faker.helpers.arrayElement([
                ActiveDirectoryNodeKind.User,
                ActiveDirectoryNodeKind.Group,
                ActiveDirectoryNodeKind.Computer,
            ]),
    };
};

export const createMockDomain = () => ({
    type: faker.helpers.arrayElement(['active-directory', 'azure']),
    impactValue: faker.datatype.number({ min: 0, max: 100 }),
    name: faker.internet.domainName(),
    id: faker.datatype.uuid(),
    collected: faker.datatype.boolean(),
});
