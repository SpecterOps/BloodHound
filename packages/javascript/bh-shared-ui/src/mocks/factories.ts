import { AssetGroup, AssetGroupMember } from 'js-client-library';
import { SearchResults } from '../hooks';

export const createAssetGroupMembers = (): { members: AssetGroupMember[] } => {
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

export const createAssetGroup = (): AssetGroup => {
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

export const createSearchResults = (): SearchResults => {
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
