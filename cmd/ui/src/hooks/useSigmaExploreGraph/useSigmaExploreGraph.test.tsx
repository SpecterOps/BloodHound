// Copyright 2025 Specter Ops, Inc.
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

import * as bhSharedUI from 'bh-shared-ui';
import { FlatGraphResponse, GraphResponse } from 'js-client-library';
import { normalizeGraphDataForSigma } from '.';

const transformToFlatGraphResponseSpy = vitest.spyOn(bhSharedUI, 'transformToFlatGraphResponse');
transformToFlatGraphResponseSpy.mockReturnValue({});

const typicalGraphResponse: GraphResponse = {
    data: {
        nodes: {},
        edges: [],
    },
};

const typicalFlatGraphResponse: FlatGraphResponse = {
    '1234': {
        color: '#DBE617',
        data: {
            admincount: false,
            description:
                'Users are prevented from making accidental or intentional system-wide changes and can run most applications',
            domain: 'FAKEDOMAIN.CORP',
            isaclprotected: false,
            lastseen: '2025-03-26T20:45:30.715175335Z',
            nodetype: 'Group',
            samaccountname: 'Users',
            whencreated: 1668394808,
            source: '',
            target: '',
            label: '',
            kind: '',
            lastSeen: '',
            objectId: '',
            isTierZero: true,
            isOwnedObject: true,
        },
        border: {
            color: 'black',
        },
        fontIcon: {
            text: 'fa-users',
        },
        label: {
            backgroundColor: 'rgba(255,255,255,0.9)',
            center: true,
            fontSize: 14,
            text: 'USERS@FAKEDOMAIN.CORP',
        },
        size: 1,
    },
};

describe('normalizeGraphDataForSigma', () => {
    it('returns undefined if graphData is undefined', () => {
        const actual = normalizeGraphDataForSigma(undefined);
        expect(actual).toBeUndefined();
    });

    it('calls transformToFlatGraphResponse when graphData matches GraphResponse interface', () => {
        normalizeGraphDataForSigma(typicalGraphResponse);
        expect(transformToFlatGraphResponseSpy).toBeCalled();
    });

    it('returns graphData as is if it doesnt match the GraphResponse interface', () => {
        normalizeGraphDataForSigma(typicalFlatGraphResponse);
        expect(transformToFlatGraphResponseSpy).not.toBeCalled();
    });
});
