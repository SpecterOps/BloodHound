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
import ExploreTable from './ExploreTable';

import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { json2csv } from 'json-2-csv';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { useState } from 'react';
import { cypherTestResponse } from '../../mocks';
import { render } from '../../test-utils';
import { makeStoreMapFromColumnOptions } from './explore-table-utils';
const SELECTED_ROW_INDICATOR_CLASS = 'shadow-[inset_0px_0px_0px_2px_var(--primary)]';

const closeCallbackSpy = vi.fn();
const kebabCallbackSpy = vi.fn();

vi.mock('js-file-download');
vi.mock('json-2-csv');

const getFirstCellOfType = (type: string) => screen.getAllByTestId(`table-cell-${type}`)[0];

const server = setupServer(
    rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
        return res(ctx.json(cypherTestResponse));
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(ctx.status(200), ctx.json({ data: [{ key: 'explore_table_view', enabled: true }] }));
    }),

    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(ctx.status(200));
    }),
    rest.get('/api/v2/custom-nodes', (req, res, ctx) => {
        return res(ctx.status(200));
    })
);

beforeAll(() => {
    Object.defineProperty(HTMLElement.prototype, 'offsetHeight', {
        value: 800,
    });
    Object.defineProperty(HTMLElement.prototype, 'offsetWidth', {
        value: 800,
    });

    server.listen();
});

const jsonToCsvArgs = [
    [
        {
            admincount: true,
            displayname: 'certman',
            distinguishedname: 'CN=CERTMAN,OU=USERS,OU=TIER0,DC=PHANTOM,DC=CORP',
            domain: 'PHANTOM.CORP',
            domainsid: 'S-1-5-21-2697957641-2271029196-387917394',
            dontreqpreauth: false,
            enabled: true,
            hasspn: false,
            isOwnedObject: false,
            isTierZero: true,
            isaclprotected: true,
            kind: 'User',
            kinds: ['User'],
            label: 'CERTMAN@PHANTOM.CORP',
            lastSeen: '2025-07-09T00:28:46.292Z',
            lastcollected: '2025-07-09T00:28:46.055264963Z',
            lastlogon: 1695966948,
            lastlogontimestamp: 1695966443,
            lastseen: '2025-07-09T00:28:46.292Z',
            name: 'CERTMAN@PHANTOM.CORP',
            objectId: 'S-1-5-21-2697957641-2271029196-387917394-2201',
            objectid: 'S-1-5-21-2697957641-2271029196-387917394-2201',
            ownersid: 'S-1-5-21-2697957641-2271029196-387917394-512',
            passwordnotreqd: false,
            pwdlastset: 1695966321,
            pwdneverexpires: true,
            samaccountname: 'certman',
            sensitive: false,
            serviceprincipalnames: [],
            sidhistory: [],
            system_tags: 'admin_tier_0',
            trustedtoauth: false,
            unconstraineddelegation: false,
            whencreated: 1695941121,
        },
        {
            admincount: true,
            displayname: 'alice',
            distinguishedname: 'CN=ALICE,OU=USERS,OU=TIER1,DC=PHANTOM,DC=CORP',
            domain: 'PHANTOM.CORP',
            domainsid: 'S-1-5-21-2697957641-2271029196-387917394',
            dontreqpreauth: false,
            enabled: true,
            hasspn: false,
            isOwnedObject: false,
            isTierZero: true,
            isaclprotected: true,
            kind: 'User',
            kinds: ['User'],
            label: 'ALICE@PHANTOM.CORP',
            lastSeen: '2025-07-09T00:28:46.292Z',
            lastcollected: '2025-07-09T00:28:46.055264963Z',
            lastlogon: 0,
            lastlogontimestamp: -1,
            lastseen: '2025-07-09T00:28:46.292Z',
            name: 'ALICE@PHANTOM.CORP',
            objectId: 'S-1-5-21-2697957641-2271029196-387917394-2173',
            objectid: 'S-1-5-21-2697957641-2271029196-387917394-2173',
            ownersid: 'S-1-5-21-2697957641-2271029196-387917394-512',
            passwordnotreqd: false,
            pwdlastset: 1681286877,
            pwdneverexpires: true,
            samaccountname: 'alice',
            sensitive: false,
            serviceprincipalnames: [],
            sidhistory: [],
            system_tags: 'admin_tier_0',
            trustedtoauth: false,
            unconstraineddelegation: false,
            whencreated: 1681261677,
        },

        {
            admincount: true,
            displayname: 'T1_TonyMontana',
            distinguishedname: 'CN=T1_TONYMONTANA,OU=USERS,OU=TIER1,DC=PHANTOM,DC=CORP',
            domain: 'PHANTOM.CORP',
            domainsid: 'S-1-5-21-2697957641-2271029196-387917394',
            dontreqpreauth: false,
            enabled: true,
            hasspn: false,
            isOwnedObject: false,
            isTierZero: false,
            isaclprotected: true,
            kind: 'User',
            kinds: ['User'],
            label: 'T1_TONYMONTANA@PHANTOM.CORP',
            lastSeen: '2025-07-09T00:28:46.306Z',
            lastcollected: '2025-07-09T00:28:46.055264963Z',
            lastlogon: 1674216447,
            lastlogontimestamp: 1673626093,
            lastseen: '2025-07-09T00:28:46.306Z',
            name: 'T1_TONYMONTANA@PHANTOM.CORP',
            objectId: 'S-1-5-21-2697957641-2271029196-387917394-2110',
            objectid: 'S-1-5-21-2697957641-2271029196-387917394-2110',
            ownersid: 'S-1-5-21-2697957641-2271029196-387917394-512',
            passwordnotreqd: false,
            pwdlastset: 1664381451,
            pwdneverexpires: true,
            samaccountname: 'T1_TonyMontana',
            sensitive: false,
            serviceprincipalnames: [],
            sidhistory: [],
            trustedtoauth: false,
            unconstraineddelegation: false,
            whencreated: 1664356251,
        },
        {
            admincount: false,
            displayname: 'zzzigne',
            distinguishedname: 'CN=ZZZIGNE,OU=USERS,OU=TIER1,DC=PHANTOM,DC=CORP',
            domain: 'PHANTOM.CORP',
            domainsid: 'S-1-5-21-2697957641-2271029196-387917394',
            dontreqpreauth: false,
            enabled: true,
            hasspn: false,
            isOwnedObject: false,
            isTierZero: false,
            isaclprotected: false,
            kind: 'User',
            kinds: ['User'],
            label: 'ZZZIGNE@PHANTOM.CORP',
            lastSeen: '2025-07-09T00:28:46.292Z',
            lastcollected: '2025-07-09T00:28:46.055264963Z',
            lastlogon: 0,
            lastlogontimestamp: -1,
            lastseen: '2025-07-09T00:28:46.292Z',
            name: 'ZZZIGNE@PHANTOM.CORP',
            objectId: 'S-1-5-21-2697957641-2271029196-387917394-2216',
            objectid: 'S-1-5-21-2697957641-2271029196-387917394-2216',
            ownersid: 'S-1-5-21-2697957641-2271029196-387917394-512',
            passwordnotreqd: false,
            pwdlastset: 1707939756,
            pwdneverexpires: true,
            samaccountname: 'zzzigne',
            sensitive: false,
            serviceprincipalnames: [],
            sidhistory: [],
            trustedtoauth: false,
            unconstraineddelegation: false,
            whencreated: 1707910956,
        },
        {
            admincount: false,
            displayname: 'svc_shs',
            distinguishedname: 'CN=SVC_SHS,OU=SHARPHOUND TEST,DC=PHANTOM,DC=CORP',
            domain: 'PHANTOM.CORP',
            domainsid: 'S-1-5-21-2697957641-2271029196-387917394',
            dontreqpreauth: false,
            enabled: true,
            hasspn: false,
            isOwnedObject: false,
            isTierZero: false,
            isaclprotected: false,
            kind: 'User',
            kinds: ['User'],
            label: 'SVC_SHS@PHANTOM.CORP',
            lastSeen: '2025-07-09T00:28:46.292Z',
            lastcollected: '2025-07-09T00:28:46.055264963Z',
            lastlogon: 0,
            lastlogontimestamp: -1,
            lastseen: '2025-07-09T00:28:46.292Z',
            name: 'SVC_SHS@PHANTOM.CORP',
            objectId: 'S-1-5-21-2697957641-2271029196-387917394-2165',
            objectid: 'S-1-5-21-2697957641-2271029196-387917394-2165',
            ownersid: 'S-1-5-21-2697957641-2271029196-387917394-512',
            passwordnotreqd: false,
            pwdlastset: 1678309565,
            pwdneverexpires: true,
            samaccountname: 'svc_shs',
            sensitive: false,
            serviceprincipalnames: [],
            sidhistory: [],
            trustedtoauth: false,
            unconstraineddelegation: false,
            whencreated: 1678280765,
        },
        {
            domain: 'PHANTOM.CORP',
            domainsid: 'S-1-5-21-2697957641-2271029196-387917394',
            isOwnedObject: false,
            isTierZero: false,
            kind: 'User',
            kinds: ['User'],
            label: 'NETWORK SERVICE@PHANTOM.CORP',
            lastSeen: '2025-07-09T00:28:46.055264963Z',
            lastcollected: '2025-07-09T00:28:46.055264963Z',
            lastseen: '2025-07-09T00:28:46.055264963Z',
            name: 'NETWORK SERVICE@PHANTOM.CORP',
            objectId: 'PHANTOM.CORP-S-1-5-20',
            objectid: 'PHANTOM.CORP-S-1-5-20',
        },
        {
            admincount: true,
            displayname: 'tom',
            distinguishedname: 'CN=TOM,CN=USERS,DC=GHOST,DC=CORP',
            domain: 'GHOST.CORP',
            domainsid: 'S-1-5-21-2845847946-3451170323-4261139666',
            dontreqpreauth: false,
            enabled: false,
            hasspn: false,
            isOwnedObject: false,
            isTierZero: true,
            isaclprotected: true,
            kind: 'User',
            kinds: ['User'],
            label: 'TOM@GHOST.CORP',
            lastSeen: '2025-07-09T00:28:46.525Z',
            lastcollected: '2025-07-09T00:28:46.504907963Z',
            lastlogon: 1352620493,
            lastlogontimestamp: 1352620493,
            lastseen: '2025-07-09T00:28:46.525Z',
            name: 'TOM@GHOST.CORP',
            objectId: 'S-1-5-21-2845847946-3451170323-4261139666-1105',
            objectid: 'S-1-5-21-2845847946-3451170323-4261139666-1105',
            ownersid: 'S-1-5-21-2845847946-3451170323-4261139666-512',
            passwordnotreqd: false,
            pwdlastset: 1294408493,
            pwdneverexpires: true,
            samaccountname: 'tom',
            sensitive: false,
            serviceprincipalnames: [],
            sidhistory: [],
            system_tags: 'admin_tier_0',
            trustedtoauth: false,
            unconstraineddelegation: false,
            whencreated: 1107163853,
        },
        {
            admincount: true,
            displayname: 'walter',
            distinguishedname: 'CN=WALTER,CN=USERS,DC=GHOST,DC=CORP',
            domain: 'GHOST.CORP',
            domainsid: 'S-1-5-21-2845847946-3451170323-4261139666',
            dontreqpreauth: false,
            enabled: false,
            hasspn: false,
            isOwnedObject: false,
            isTierZero: true,
            isaclprotected: true,
            kind: 'User',
            kinds: ['User'],
            label: 'WALTER@GHOST.CORP',
            lastSeen: '2025-07-09T00:28:46.525Z',
            lastcollected: '2025-07-09T00:28:46.504907963Z',
            lastlogon: 1221657838,
            lastlogontimestamp: 1221657838,
            lastseen: '2025-07-09T00:28:46.525Z',
            name: 'WALTER@GHOST.CORP',
            objectId: 'S-1-5-21-2845847946-3451170323-4261139666-1106',
            objectid: 'S-1-5-21-2845847946-3451170323-4261139666-1106',
            ownersid: 'S-1-5-21-2845847946-3451170323-4261139666-512',
            passwordnotreqd: false,
            pwdlastset: 1117877458,
            pwdneverexpires: true,
            samaccountname: 'walter',
            sensitive: false,
            serviceprincipalnames: [],
            sidhistory: [],
            system_tags: 'admin_tier_0',
            trustedtoauth: false,
            unconstraineddelegation: false,
            whencreated: 1107164693,
        },
        {
            admincount: true,
            description: 'Built-in account for administering the computer/domain',
            distinguishedname: 'CN=ADMINISTRATOR,CN=USERS,DC=GHOST,DC=CORP',
            domain: 'GHOST.CORP',
            domainsid: 'S-1-5-21-2845847946-3451170323-4261139666',
            dontreqpreauth: false,
            enabled: true,
            hasspn: false,
            isOwnedObject: false,
            isTierZero: true,
            isaclprotected: true,
            kind: 'User',
            kinds: ['User'],
            label: 'ADMINISTRATOR@GHOST.CORP',
            lastSeen: '2025-07-09T00:28:46.525Z',
            lastcollected: '2025-07-09T00:28:46.504907963Z',
            lastlogon: 1702913193,
            lastlogontimestamp: 1702913054,
            lastseen: '2025-07-09T00:28:46.525Z',
            name: 'ADMINISTRATOR@GHOST.CORP',
            objectId: 'S-1-5-21-2845847946-3451170323-4261139666-500',
            objectid: 'S-1-5-21-2845847946-3451170323-4261139666-500',
            ownersid: 'S-1-5-21-2845847946-3451170323-4261139666-512',
            passwordnotreqd: false,
            pwdlastset: 1674763808,
            pwdneverexpires: true,
            samaccountname: 'Administrator',
            sensitive: false,
            serviceprincipalnames: [],
            sidhistory: [],
            system_tags: 'admin_tier_0',
            trustedtoauth: false,
            unconstraineddelegation: false,
            whencreated: 1656555664,
        },
        {
            admincount: false,
            description: 'Built-in account for guest access to the computer/domain',
            distinguishedname: 'CN=GUEST,CN=USERS,DC=GHOST,DC=CORP',
            domain: 'GHOST.CORP',
            domainsid: 'S-1-5-21-2845847946-3451170323-4261139666',
            dontreqpreauth: false,
            enabled: false,
            hasspn: false,
            isOwnedObject: false,
            isTierZero: false,
            isaclprotected: false,
            kind: 'User',
            kinds: ['User'],
            label: 'GUEST@GHOST.CORP',
            lastSeen: '2025-07-09T00:28:46.525Z',
            lastcollected: '2025-07-09T00:28:46.504907963Z',
            lastlogon: 0,
            lastlogontimestamp: -1,
            lastseen: '2025-07-09T00:28:46.525Z',
            name: 'GUEST@GHOST.CORP',
            objectId: 'S-1-5-21-2845847946-3451170323-4261139666-501',
            objectid: 'S-1-5-21-2845847946-3451170323-4261139666-501',
            ownersid: 'GHOST.CORP-S-1-5-32-544',
            passwordnotreqd: true,
            pwdlastset: 0,
            pwdneverexpires: true,
            samaccountname: 'Guest',
            sensitive: false,
            serviceprincipalnames: [],
            sidhistory: [],
            trustedtoauth: false,
            unconstraineddelegation: false,
            whencreated: 1656555664,
        },
    ],
    {
        emptyFieldValue: '',
        preventCsvInjection: true,
        keys: [
            'admincount',
            'description',
            'displayname',
            'distinguishedname',
            'domain',
            'domainsid',
            'dontreqpreauth',
            'email',
            'enabled',
            'gmsa',
            'hasspn',
            'isaclprotected',
            'lastcollected',
            'lastlogon',
            'lastlogontimestamp',
            'msa',
            'ownersid',
            'passwordnotreqd',
            'pwdlastset',
            'pwdneverexpires',
            'samaccountname',
            'sensitive',
            'serviceprincipalnames',
            'sidhistory',
            'system_tags',
            'title',
            'trustedtoauth',
            'unconstraineddelegation',
            'whencreated',
            'kind',
            'label',
            'objectId',
            'isTierZero',
            'isOwnedObject',
            'lastSeen',
        ],
    },
];

const WrappedExploreTable = () => {
    const [selectedColumns, setSelectedColumns] = useState<Record<string, boolean>>({
        kind: true,
        isTierZero: true,
        label: true,
        objectId: true,
    });

    return (
        <ExploreTable
            open={true}
            selectedColumns={selectedColumns}
            onManageColumnsChange={(columns) => {
                const newColumns = makeStoreMapFromColumnOptions(columns);
                setSelectedColumns(newColumns);
            }}
            onClose={closeCallbackSpy}
            onKebabMenuClick={(row) => {
                kebabCallbackSpy(row);
            }}
        />
    );
};

const setup = async () => {
    render(<WrappedExploreTable />, { route: `/graphview?searchType=cypher&cypherSearch=encodedquery` });

    const user = userEvent.setup();

    return { user };
};

describe('ExploreTable', async () => {
    it('should render', async () => {
        await setup();

        expect(await screen.findByText('10 results')).toBeInTheDocument();
        expect(screen.getByText('Object ID')).toBeInTheDocument();
        expect(screen.getByText('Node Type')).toBeInTheDocument();
        expect(screen.getByText('Name')).toBeInTheDocument();
        expect(screen.getByText('CERTMAN@PHANTOM.CORP')).toBeInTheDocument();
        expect(screen.getByText('S-1-5-21-2697957641-2271029196-387917394-2201')).toBeInTheDocument();
        expect(screen.queryByText('Domain FQDN')).not.toBeInTheDocument();
    });
    it('"Columns" button allows user to edit column settings', async () => {
        const { user } = await setup();

        // await for results to return
        await screen.findByText('10 results');

        const manageColumnsButton = screen.getByRole('button', { name: 'Columns' });
        expect(screen.queryByText('Domain FQDN')).not.toBeInTheDocument();
        expect(screen.queryByText('Admin Count')).not.toBeInTheDocument();

        await user.click(manageColumnsButton);

        expect(await screen.findByText('Admin Count')).toBeInTheDocument();
        const domainListItem = screen.getByText('Domain FQDN');
        expect(domainListItem).toBeInTheDocument();

        expect(screen.getByRole('listbox')).toBeInTheDocument();
        await user.click(domainListItem);

        expect(screen.getAllByText('Domain FQDN')).toHaveLength(2);

        // click anywhere outside the combobox dropdown to close manage columns component
        await user.click(screen.getByText('Results'));

        // Demonstrates that combobox is closed
        expect(screen.queryByText('Admin Count')).not.toBeInTheDocument();
        expect(screen.getByText('Domain FQDN')).toBeInTheDocument();
    });

    it('Clicking header allows user to sort by column', async () => {
        const { user } = await setup();

        await screen.findByText('10 results');

        // Unsorted first display name cell
        expect(getFirstCellOfType('label')).toHaveTextContent('CERTMAN@PHANTOM.CORP');

        // Alphabetically sorted first display name cell
        await user.click(screen.getByText('Name'));
        expect(getFirstCellOfType('label')).toHaveTextContent('ADMINISTRATOR@GHOST.CORP');

        // Reverse Alphabetically sorted first display name cell
        await user.click(screen.getByText('Name'));
        expect(getFirstCellOfType('label')).toHaveTextContent('ZZZIGNE@PHANTOM.CORP');

        // Reset to unsorted
        await user.click(screen.getByText('Name'));
        expect(getFirstCellOfType('label')).toHaveTextContent('CERTMAN@PHANTOM.CORP');

        // Unsorted first object id cell
        expect(getFirstCellOfType('objectId')).toHaveTextContent('S-1-5-21-2697957641-2271029196-387917394-2201');

        // Descending sorted first object id cell
        await user.click(screen.getByText('Object ID'));
        expect(getFirstCellOfType('objectId')).toHaveTextContent('PHANTOM.CORP-S-1-5-20');

        // Ascending sorted first object id cell
        await user.click(screen.getByText('Object ID'));
        expect(getFirstCellOfType('objectId')).toHaveTextContent('S-1-5-21-2845847946-3451170323-4261139666-1106');

        // Reset to unsorted
        await user.click(screen.getByText('Object ID'));
        expect(getFirstCellOfType('objectId')).toHaveTextContent('S-1-5-21-2697957641-2271029196-387917394-2201');
    });

    it('Expand button causes table to expand to full height', async () => {
        const { user } = await setup();

        const container = screen.getByTestId('explore-table-container-wrapper');
        expect(container.className).toContain('h-1/2');
        const expandButton = screen.getByTestId('expand-button');

        await user.click(expandButton);

        expect(container.className).toContain('h-[calc(100%');
    });

    it('Download button causes the json2csv function to be called', async () => {
        const { user } = await setup();

        expect(json2csv).not.toBeCalled();
        const downloadButton = screen.getByTestId('download-button');

        await user.click(downloadButton);

        expect(json2csv).toBeCalledWith(...jsonToCsvArgs);
    });

    it('Close button click causes the callback function to be called', async () => {
        const { user } = await setup();

        expect(closeCallbackSpy).not.toBeCalled();
        const closeButton = screen.getByTestId('close-button');

        await user.click(closeButton);

        expect(closeCallbackSpy).toBeCalled();
    });

    it('Typing in the search bar filters the results', async () => {
        const { user } = await setup();

        const name_1 = 'ZZZIGNE@PHANTOM.CORP';
        const object_id_of_name_1 = 'S-1-5-21-2697957641-2271029196-387917394-2216';
        const name_2 = 'WALTER@GHOST.CORP';
        const name_3 = 'CERTMAN@PHANTOM.CORP';
        const searchInput = screen.getByTestId('explore-table-search');

        const andyRowBefore = await screen.findByText(name_1);
        const svcRowBefore = screen.getByText(name_2);
        const guestRowBefore = screen.getByText(name_3);

        expect(andyRowBefore).toBeInTheDocument();
        expect(guestRowBefore).toBeInTheDocument();
        expect(svcRowBefore).toBeInTheDocument();

        await user.type(searchInput, object_id_of_name_1);

        const andyRowAfter = screen.queryByText(name_1);
        const svcRowAfter = screen.queryByText(name_2);
        const guestRowAfter = screen.queryByText(name_3);

        expect(andyRowAfter).toBeInTheDocument();
        expect(guestRowAfter).not.toBeInTheDocument();
        expect(svcRowAfter).not.toBeInTheDocument();

        await user.clear(searchInput);
        await user.type(searchInput, name_3);

        const andyRowFinal = screen.queryByText(name_1);
        const svcRowFinal = screen.queryByText(name_2);
        const guestRowFinal = screen.queryByText(name_3);

        expect(andyRowFinal).not.toBeInTheDocument();
        expect(guestRowFinal).toBeInTheDocument();
        expect(svcRowFinal).not.toBeInTheDocument();
    });

    it('Clicking on a row causes row to be selected', async () => {
        const { user } = await setup();

        const jdPhantomRow = await screen.findByRole('row', { name: /TOM@GHOST.CORP/ });

        expect(jdPhantomRow.className).not.toContain(SELECTED_ROW_INDICATOR_CLASS);

        await user.click(jdPhantomRow);

        expect(kebabCallbackSpy).not.toBeCalled();

        expect(jdPhantomRow.className).toContain(SELECTED_ROW_INDICATOR_CLASS);
    });

    it('Kebab menu click causes the callback function to be called with the correct parameters', async () => {
        const { user } = await setup();

        expect(kebabCallbackSpy).not.toBeCalled();

        const jdPhantomRow = await screen.findByRole('row', { name: /TOM@GHOST.CORP/ });

        const kebabButton = within(jdPhantomRow).getByTestId('kebab-menu');

        await user.click(kebabButton);

        expect(kebabCallbackSpy).toBeCalledWith({
            id: '569',
            x: 0,
            y: 0,
        });
    });
});
