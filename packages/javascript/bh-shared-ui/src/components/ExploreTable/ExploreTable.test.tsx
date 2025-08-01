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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { useState } from 'react';
import { cypherTestResponse } from '../../mocks';
import { render } from '../../test-utils';
import * as exportUtils from '../../utils/exportGraphData';
import { makeStoreMapFromColumnOptions } from './explore-table-utils';

const SELECTED_ROW_INDICATOR_CLASS = 'shadow-[inset_0px_0px_0px_2px_var(--primary)]';

const exportToJsonSpy = vi.spyOn(exportUtils, 'exportToJson');
exportToJsonSpy.mockImplementation(() => undefined);

const closeCallbackSpy = vi.fn();
const kebabCallbackSpy = vi.fn();

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
            onKebabMenuClick={kebabCallbackSpy}
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
    it('Manage Columns allow user to change columns', async () => {
        const { user } = await setup();

        // await for results to return
        await screen.findByText('10 results');

        const manageColumnsButton = screen.getByRole('button', { name: 'Manage Columns' });
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

        await user.click(screen.getByText('Name'));

        // Alphabetically sorted first display name cell
        expect(getFirstCellOfType('label')).toHaveTextContent('ADMINISTRATOR@GHOST.CORP');

        await user.click(screen.getByText('Name'));

        // Reverse Alphabetically sorted first display name cell
        expect(getFirstCellOfType('label')).toHaveTextContent('ZZZIGNE@PHANTOM.CORP');

        await user.click(screen.getByText('Object ID'));

        // Descending sorted first object id cell
        expect(getFirstCellOfType('objectId')).toHaveTextContent('PHANTOM.CORP-S-1-5-20');

        await user.click(screen.getByText('Object ID'));

        // Ascending sorted first object id cell
        expect(getFirstCellOfType('objectId')).toHaveTextContent('S-1-5-21-2845847946-3451170323-4261139666-1106');
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

        expect(exportToJsonSpy).not.toBeCalled();
        const downloadButton = screen.getByTestId('download-button');

        await user.click(downloadButton);

        expect(exportToJsonSpy).toBeCalledWith({ nodes: cypherTestResponse.data.nodes });
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

        expect(jdPhantomRow.className).not.toContain(SELECTED_ROW_INDICATOR_CLASS);

        await user.click(kebabButton);

        expect(kebabCallbackSpy).toBeCalledWith({
            id: '569',
            x: 0,
            y: 0,
        });

        expect(jdPhantomRow.className).toContain(SELECTED_ROW_INDICATOR_CLASS);
    });
});
