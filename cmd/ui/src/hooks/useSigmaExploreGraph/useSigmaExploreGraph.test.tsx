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
