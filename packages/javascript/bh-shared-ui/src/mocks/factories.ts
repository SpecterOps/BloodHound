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

import { AssetGroup, AssetGroupMember, AssetGroupMemberParams } from 'js-client-library';
import { SearchResults } from '../hooks';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '..';

export const createMockAssetGroupMembers = (): { members: AssetGroupMember[] } => {
    return {
        members: [
            {
                asset_group_id: 1,
                object_id: '00000-00001',
                primary_kind: 'User',
                kinds: ['User', 'Base'],
                environment_id: '00000-00000-00001',
                environment_kind: 'Domain',
                name: 'USER_00001@TESTLAB.LOCAL',
                custom_member: false,
            },
            {
                asset_group_id: 1,
                object_id: '00000-00002',
                primary_kind: 'Computer',
                kinds: ['Computer', 'Base'],
                environment_id: '00000-00000-00001',
                environment_kind: 'Domain',
                name: 'COMPUTER_00001@TESTLAB.LOCAL',
                custom_member: false,
            },
            {
                asset_group_id: 1,
                object_id: '00000-00003',
                primary_kind: 'GPO',
                kinds: ['GPO', 'Base'],
                environment_id: '00000-00000-00001',
                environment_kind: 'Domain',
                name: 'GPO_00001@TESTLAB.LOCAL',
                custom_member: true,
            },
        ],
    };
};

export const createMockAssetGroup = (): AssetGroup => {
    return {
        id: 1,
        name: 'Admin Tier Zero',
        tag: 'admin_tier_0',
        member_count: 3,
        system_group: true,
        Selectors: [],
        created_at: '2023-10-18T16:19:25.26533Z',
        updated_at: '2023-10-18T16:19:25.26533Z',
        deleted_at: {
            Time: '0001-01-01T00:00:00Z',
            Valid: false,
        },
    };
};

export const createMockSearchResults = (): SearchResults => {
    return [
        {
            objectid: '00000-00000-00000-00001',
            type: 'Computer',
            name: '00001.TESTLAB.LOCAL',
            distinguishedname: '',
            system_tags: '',
        },
    ];
};

export const createMockAssetGroupMemberParams = (): AssetGroupMemberParams => {
    return {
        environment_id: '000-000-000',
        primary_kind: 'eq:Domain',
        custom_member: 'eq:true'
    }
}

export const createMockAvailableNodeKinds = (): Array<ActiveDirectoryNodeKind | AzureNodeKind> => {
    return [
        ActiveDirectoryNodeKind.User,
        ActiveDirectoryNodeKind.Computer,
        ActiveDirectoryNodeKind.Domain,
    ]
}