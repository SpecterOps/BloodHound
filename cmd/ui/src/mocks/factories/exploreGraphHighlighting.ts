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

import { MultiDirectedGraph } from 'graphology';

const testNodes = {
    '51154': {
        color: '#DBE617',
        data: {
            isTierZero: false,
            kinds: ['Base', 'Group'],
            name: 'AUTHENTICATED USERS@WRAITH.CORP',
            nodetype: 'Group',
            objectid: 'WRAITH.CORP-S-1-5-11',
            system_tags: null,
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-users' },
        label: {
            backgroundColor: 'rgba(255,255,255,0.9)',
            center: true,
            fontSize: 14,
            text: 'AUTHENTICATED USERS@WRAITH.CORP',
        },
        size: 1,
    },
    '51155': {
        color: '#DBE617',
        data: {
            isTierZero: false,
            kinds: ['Base', 'Group'],
            name: 'DOMAIN USERS@WRAITH.CORP',
            nodetype: 'Group',
            objectid: 'S-1-5-21-3702535222-3822678775-2090119576-513',
            system_tags: null,
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-users' },
        label: {
            backgroundColor: 'rgba(255,255,255,0.9)',
            center: true,
            fontSize: 14,
            text: 'DOMAIN USERS@WRAITH.CORP',
        },
        size: 1,
    },
    '52074': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'USERNOEMAILNOSEC@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '22A32FDF-278D-44C2-8000-1F2AEECC117F',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: {
            backgroundColor: 'rgba(255,255,255,0.9)',
            center: true,
            fontSize: 14,
            text: 'USERNOEMAILNOSEC@WRAITH.CORP',
        },
        size: 1,
    },
    '52097': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'USER@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '0059CED1-7362-4506-AD56-DAEEFAAE8628',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'USER@WRAITH.CORP' },
        size: 1,
    },
    '52159': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'CLIENTAUTH@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '9CC41FA5-7C15-4D50-949C-38C17BEB6F7C',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'CLIENTAUTH@WRAITH.CORP' },
        size: 1,
    },
    '52172': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'EFS@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '7E43AF6D-951D-48E7-8CC0-6546742A222F',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'EFS@WRAITH.CORP' },
        size: 1,
    },
    '52192': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'USEREMAILNOSEC@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '035A964D-8533-4110-8A98-E3CAF15920EC',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: {
            backgroundColor: 'rgba(255,255,255,0.9)',
            center: true,
            fontSize: 14,
            text: 'USEREMAILNOSEC@WRAITH.CORP',
        },
        size: 1,
    },
    '52254': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'USEREMAIL@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: 'DA2D7CF7-325E-4029-8168-18A0EACF2A0C',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'USEREMAIL@WRAITH.CORP' },
        size: 1,
    },
    '52270': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'USERV2@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '545E2256-6CE8-4EC8-8196-C7A180243EFF',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'USERV2@WRAITH.CORP' },
        size: 1,
    },
    '52278': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'USERNOUPN@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '6685875C-5565-460B-822F-57586AF0AB76',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'USERNOUPN@WRAITH.CORP' },
        size: 1,
    },
    '52314': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'EXTMAIL@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '59B53E94-6DD2-4EC4-AB97-576327BACFEE',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'EXTMAIL@WRAITH.CORP' },
        size: 1,
    },
    '52324': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'USERSIGNATURE@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '5EFC7529-63B5-4524-8EFC-8B99A8275CED',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: {
            backgroundColor: 'rgba(255,255,255,0.9)',
            center: true,
            fontSize: 14,
            text: 'USERSIGNATURE@WRAITH.CORP',
        },
        size: 1,
    },
    '52339': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'COMPUTERV2@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '9FD91BB3-DB77-4509-B458-6F5D2E6EBF7F',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'COMPUTERV2@WRAITH.CORP' },
        size: 1,
    },
    '52350': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'USERNOSANFLAGS@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '13AA88AA-7AEA-4325-A9EA-55A12AB45FD2',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: {
            backgroundColor: 'rgba(255,255,255,0.9)',
            center: true,
            fontSize: 14,
            text: 'USERNOSANFLAGS@WRAITH.CORP',
        },
        size: 1,
    },
    '52357': {
        color: '#B153F3',
        data: {
            isTierZero: true,
            kinds: ['Base', 'CertTemplate', 'Tag_Tier_Zero'],
            level: 0,
            name: 'MYTEMPLATE@WRAITH.CORP',
            nodetype: 'CertTemplate',
            objectid: '73473DD3-86B2-494C-8A92-8E214873B890',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-id-card' },
        label: { backgroundColor: 'rgba(255,255,255,0.9)', center: true, fontSize: 14, text: 'MYTEMPLATE@WRAITH.CORP' },
        size: 1,
    },
    '52952': {
        color: '#4696E9',
        data: {
            isTierZero: true,
            kinds: ['Base', 'EnterpriseCA', 'Tag_Tier_Zero'],
            level: 0,
            name: 'WRAITH-EXTCA02-CA@WRAITH.CORP',
            nodetype: 'EnterpriseCA',
            objectid: '396FFDC9-079C-4AB1-9816-ACEED046BCDD',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-building' },
        label: {
            backgroundColor: 'rgba(255,255,255,0.9)',
            center: true,
            fontSize: 14,
            text: 'WRAITH-EXTCA02-CA@WRAITH.CORP',
        },
        size: 1,
    },
    '53069': {
        color: '#17E625',
        data: {
            isTierZero: false,
            kinds: ['Base', 'User'],
            name: '$N31000-BBHT9IVGI4SP@WRAITH.CORP',
            nodetype: 'User',
            objectid: 'S-1-5-21-3702535222-3822678775-2090119576-1143',
            system_tags: null,
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-user' },
        label: {
            backgroundColor: 'rgba(255,255,255,0.9)',
            center: true,
            fontSize: 14,
            text: '$N31000-BBHT9IVGI4SP@WRAITH.CORP',
        },
        size: 1,
    },
    '53123': {
        color: '#4696E9',
        data: {
            isTierZero: true,
            kinds: ['Base', 'EnterpriseCA', 'Tag_Tier_Zero'],
            level: 0,
            name: 'WRAITH-EXTCA01-CA@WRAITH.CORP',
            nodetype: 'EnterpriseCA',
            objectid: '4F334FEC-0BCE-46A0-AF60-F1C481829440',
            system_tags: 'admin_tier_0',
        },
        border: { color: 'black' },
        fontIcon: { text: 'fas fa-building' },
        label: {
            backgroundColor: 'rgba(255,255,255,0.9)',
            center: true,
            fontSize: 14,
            text: 'WRAITH-EXTCA01-CA@WRAITH.CORP',
        },
        size: 1,
    },
};

const testEdges = {
    rel_925823: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '51154', label: { text: 'MemberOf' } },
    rel_931961: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52350', label: { text: 'Enroll' } },
    rel_932069: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52278', label: { text: 'Enroll' } },
    rel_932095: { color: '3a5464', end2: { arrow: true }, id1: '51154', id2: '52339', label: { text: 'Enroll' } },
    rel_932333: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52172', label: { text: 'Enroll' } },
    rel_932768: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52270', label: { text: 'Enroll' } },
    rel_932820: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52074', label: { text: 'Enroll' } },
    rel_932851: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52097', label: { text: 'Enroll' } },
    rel_932920: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52324', label: { text: 'Enroll' } },
    rel_933085: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52254', label: { text: 'Enroll' } },
    rel_933288: { color: '3a5464', end2: { arrow: true }, id1: '51154', id2: '52270', label: { text: 'Enroll' } },
    rel_933310: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52314', label: { text: 'Enroll' } },
    rel_933472: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52159', label: { text: 'Enroll' } },
    rel_933583: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52192', label: { text: 'Enroll' } },
    rel_933680: { color: '3a5464', end2: { arrow: true }, id1: '51155', id2: '52357', label: { text: 'Enroll' } },
    rel_937904: { color: '3a5464', end2: { arrow: true }, id1: '51154', id2: '52952', label: { text: 'Enroll' } },
    rel_938942: { color: '3a5464', end2: { arrow: true }, id1: '51154', id2: '53123', label: { text: 'Enroll' } },
    rel_939292: { color: '3a5464', end2: { arrow: true }, id1: '53069', id2: '51155', label: { text: 'MemberOf' } },
};

export const buildTestGraph = () => {
    const graph = new MultiDirectedGraph();
    for (const [key] of Object.entries(testNodes)) {
        graph.addNode(key);
    }
    for (const [key, value] of Object.entries(testEdges)) {
        graph.addDirectedEdgeWithKey(key, value.id1, value.id2);
    }
    return graph;
};
