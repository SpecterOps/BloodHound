import { rest } from 'msw';
import { setupServer } from 'msw/node';
import ExploreTable from './ExploreTable';

import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { render } from '../../test-utils';
import { mockCodemirrorLayoutMethods } from '../../utils';
import exploreTableTestProps from './explore-table-test-props';
import { makeStoreMapFromColumnOptions } from './explore-table-utils';

const SELECTED_ROW_INDICATOR_CLASS = 'border-primary';

const server = setupServer(
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [{ id: 1, key: 'tier_management_engine', enabled: true }],
            })
        );
    }),
    rest.get('/api/v2/graphs/kinds', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: { kinds: ['Tier Zero', 'Tier One', 'Tier Two'] },
            })
        );
    }),
    rest.get(`/api/v2/custom-nodes`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    })
);

beforeAll(() => server.listen());
beforeEach(() => mockCodemirrorLayoutMethods());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const downloadCallbackSpy = vi.fn();
const closeCallbackSpy = vi.fn();
const kebabCallbackSpy = vi.fn();

const WrappedExploreTable = () => {
    const [selectedColumns, setSelectedColumns] = useState<Record<string, boolean>>();
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
        expect(screen.getByText('Type')).toBeInTheDocument();
        expect(screen.getByText('Display Name')).toBeInTheDocument();
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

        const tableCells = screen.getAllByRole('cell');

        // A bit brittle, but as long as the component remains 'table shaped', this should work
        const firstDisplayNameCell = tableCells[3];

        // Unsorted first display name cell
        expect(firstDisplayNameCell).toHaveTextContent('CERTMAN@PHANTOM.CORP');

        await user.click(screen.getByText('Display Name'));

        // Alphabetically sorted first display name cell
        expect(firstDisplayNameCell).toHaveTextContent('ALICE@PHANTOM.CORP');

        await user.click(screen.getByText('Display Name'));

        // Reverse Alphabetically sorted first display name cell
        expect(firstDisplayNameCell).toHaveTextContent('T1_TONYMONTANA@PHANTOM.CORP');

        await user.click(screen.getByText('Object ID'));

        const firstObjectIdCell = tableCells[2];

        // Descending sorted first object id cell
        expect(firstObjectIdCell).toHaveTextContent('S-1-5-21-2697957641-2271029196-387917394-2110');

        await user.click(screen.getByText('Object ID'));

        // Ascending sorted first object id cell
        expect(firstObjectIdCell).toHaveTextContent('S-1-5-21-2697957641-2271029196-387917394-501');
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

    it('Clicking on a raw causes row to be selected', async () => {
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
