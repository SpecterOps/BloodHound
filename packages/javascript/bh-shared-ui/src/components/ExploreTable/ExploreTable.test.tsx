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
import { useState } from 'react';
import { render } from '../../test-utils';
import exploreTableTestProps from './explore-table-test-props';
import { makeStoreMapFromColumnOptions } from './explore-table-utils';

const SELECTED_ROW_INDICATOR_CLASS = 'border-primary';

const downloadCallbackSpy = vi.fn();
const closeCallbackSpy = vi.fn();
const kebabCallbackSpy = vi.fn();

const getFirstCellOfType = (type: string) => screen.getAllByTestId(`table-cell-${type}`)[0];

const WrappedExploreTable = () => {
    const [selectedColumns, setSelectedColumns] = useState<Record<string, boolean>>({
        nodetype: true,
        isTierZero: true,
        name: true,
        objectid: true,
    });
    const [selectedNode, setSelectedNode] = useState<string>();

    return (
        <ExploreTable
            {...exploreTableTestProps}
            selectedColumns={selectedColumns}
            onManageColumnsChange={(columns) => {
                const newColumns = makeStoreMapFromColumnOptions(columns);
                setSelectedColumns(newColumns);
            }}
            onDownloadClick={downloadCallbackSpy}
            onClose={closeCallbackSpy}
            selectedNode={selectedNode || null}
            onKebabMenuClick={kebabCallbackSpy}
            onRowClick={(row) => setSelectedNode(row.id)}
        />
    );
};

const setup = async () => {
    render(<WrappedExploreTable />);

    const user = userEvent.setup();

    return { user };
};

describe('ExploreTable', async () => {
    it('should render', async () => {
        await setup();
        expect(screen.getByText('10 results')).toBeInTheDocument();
        expect(screen.getByText('Object ID')).toBeInTheDocument();
        expect(screen.getByText('Nodetype')).toBeInTheDocument();
        expect(screen.getByText('Name')).toBeInTheDocument();
        expect(screen.getByText('CERTMAN@PHANTOM.CORP')).toBeInTheDocument();
        expect(screen.getByText('S-1-5-21-2697957641-2271029196-387917394-2201')).toBeInTheDocument();
        expect(screen.queryByText('Domain FQDN')).not.toBeInTheDocument();
    });
    it('Manage Columns allow user to change columns', async () => {
        const { user } = await setup();

        const manageColumnsButton = screen.getByRole('button', { name: 'Manage Columns' });
        expect(screen.queryByText('Domain FQDN')).not.toBeInTheDocument();
        expect(screen.queryByText('Admin Count')).not.toBeInTheDocument();

        await user.click(manageColumnsButton);

        expect(screen.getByText('Admin Count')).toBeInTheDocument();
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

        // Unsorted first display name cell
        expect(getFirstCellOfType('name')).toHaveTextContent('CERTMAN@PHANTOM.CORP');

        await user.click(screen.getByText('Name'));

        // Alphabetically sorted first display name cell
        expect(getFirstCellOfType('name')).toHaveTextContent('ALICE@PHANTOM.CORP');

        await user.click(screen.getByText('Name'));

        // Reverse Alphabetically sorted first display name cell
        expect(getFirstCellOfType('name')).toHaveTextContent('T1_TONYMONTANA@PHANTOM.CORP');

        await user.click(screen.getByText('Object ID'));

        // Descending sorted first object id cell
        expect(getFirstCellOfType('objectid')).toHaveTextContent('S-1-5-21-2697957641-2271029196-387917394-2110');

        await user.click(screen.getByText('Object ID'));

        // Ascending sorted first object id cell
        expect(getFirstCellOfType('objectid')).toHaveTextContent('S-1-5-21-2697957641-2271029196-387917394-501');
    });

    it('Expand button causes table to expand to full height', async () => {
        const { user } = await setup();

        const container = screen.getByTestId('explore-table-container-wrapper');
        expect(container.className).toContain('h-1/2');
        const expandButton = screen.getByTestId('expand-button');

        await user.click(expandButton);

        expect(container.className).toContain('h-[calc(100%');
    });

    it('Download button causes the callback function to be called', async () => {
        const { user } = await setup();

        expect(downloadCallbackSpy).not.toBeCalled();
        const downloadButton = screen.getByTestId('download-button');

        await user.click(downloadButton);

        expect(downloadCallbackSpy).toBeCalled();
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

        const SVC_DOMAIN_NAME = 'SVC_DOMAINJOIN@PHANTOM.CORP';
        const ANDY_OBJECT_ID = 'S-1-5-21-2697957641-2271029196-387917394-2187';
        const ANDY_NAME = 'ANDY@PHANTOM.CORP';
        const GUEST_NAME = 'GUEST@PHANTOM.CORP';
        const searchInput = screen.getByTestId('explore-table-search');

        const andyRowBefore = screen.getByText(ANDY_NAME);
        const svcRowBefore = screen.getByText(SVC_DOMAIN_NAME);
        const guestRowBefore = screen.getByText(GUEST_NAME);

        expect(andyRowBefore).toBeInTheDocument();
        expect(guestRowBefore).toBeInTheDocument();
        expect(svcRowBefore).toBeInTheDocument();

        await user.type(searchInput, ANDY_OBJECT_ID);

        const andyRowAfter = screen.queryByText(ANDY_NAME);
        const svcRowAfter = screen.queryByText(SVC_DOMAIN_NAME);
        const guestRowAfter = screen.queryByText(GUEST_NAME);

        expect(andyRowAfter).toBeInTheDocument();
        expect(guestRowAfter).not.toBeInTheDocument();
        expect(svcRowAfter).not.toBeInTheDocument();

        await user.clear(searchInput);
        await user.type(searchInput, GUEST_NAME);

        const andyRowFinal = screen.queryByText(ANDY_NAME);
        const svcRowFinal = screen.queryByText(SVC_DOMAIN_NAME);
        const guestRowFinal = screen.queryByText(GUEST_NAME);

        expect(andyRowFinal).not.toBeInTheDocument();
        expect(guestRowFinal).toBeInTheDocument();
        expect(svcRowFinal).not.toBeInTheDocument();
    });

    it('Clicking on a row causes row to be selected', async () => {
        const { user } = await setup();

        const jdPhantomRow = screen.getByRole('row', { name: /JD@PHANTOM.CORP/ });

        expect(jdPhantomRow.className).not.toContain(SELECTED_ROW_INDICATOR_CLASS);

        await user.click(jdPhantomRow);

        expect(kebabCallbackSpy).not.toBeCalled();

        expect(jdPhantomRow.className).toContain(SELECTED_ROW_INDICATOR_CLASS);
    });

    it('Kebab menu click causes the callback function to be called with the correct parameters', async () => {
        const { user } = await setup();

        expect(kebabCallbackSpy).not.toBeCalled();

        const jdPhantomRow = screen.getByRole('row', { name: /JD@PHANTOM.CORP/ });

        const kebabButton = within(jdPhantomRow).getByTestId('kebab-menu');

        expect(jdPhantomRow.className).not.toContain(SELECTED_ROW_INDICATOR_CLASS);

        await user.click(kebabButton);

        expect(kebabCallbackSpy).toBeCalledWith({
            id: '112',
            x: 0,
            y: 0,
        });

        expect(jdPhantomRow.className).toContain(SELECTED_ROW_INDICATOR_CLASS);
    });
});
