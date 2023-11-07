import { setupServer } from 'msw/node';
import { act, fireEvent, render, waitFor } from '../../test-utils';
import GroupManagement from './GroupManagement';
import { rest } from 'msw';
import { createMockDomain } from 'src/mocks/factories';
import { createMockAssetGroup, createMockAssetGroupMembers } from 'bh-shared-ui';

const domain = createMockDomain();
const assetGroup = createMockAssetGroup();
const assetGroupMembers = createMockAssetGroupMembers();

const server = setupServer(
    rest.get('/api/v2/available-domains', (req, res, ctx) => {
        return res(ctx.json({ data: [domain] }));
    }),
    rest.get('/api/v2/asset-groups', (req, res, ctx) => {
        return res(ctx.json({ data: { asset_groups: [assetGroup] } }));
    }),
    rest.get('/api/v2/asset-groups/1/members', (req, res, ctx) => {
        return res(
            ctx.json({
                count: assetGroupMembers.members.length,
                limit: 100,
                skip: 0,
                data: assetGroupMembers,
            })
        );
    }),
    rest.get('*', (req, res, ctx) => res(ctx.json({ data: [] })))
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('GroupManagement', () => {
    const setup = async () => await act(async () => render(<GroupManagement />));

    it('renders group and tenant dropdown selectors', async () => {
        const screen = await setup();
        const groupSelector = screen.getByTestId('dropdown_context-selector');
        const tenantSelector = await waitFor(() => screen.getByTestId('data-quality_context-selector'));

        expect(screen.getByText('Group:')).toBeInTheDocument();
        expect(screen.getByText('Tenant:')).toBeInTheDocument();
        expect(groupSelector).toBeInTheDocument();
        expect(tenantSelector).toBeInTheDocument();
    });

    it('renders an edit form for the selected asset group', async () => {
        const screen = await setup();
        const input = screen.getByRole('combobox');
        expect(input).toBeInTheDocument();
    });

    it('renders a list of asset group members', async () => {
        const screen = await setup();
        const member = assetGroupMembers.members[0];

        expect(screen.getByRole('table')).toBeInTheDocument();
        expect(screen.getByText(member.name)).toBeInTheDocument();
    });

    it('renders an empty message for the entity panel before a node is selected', async () => {
        const screen = await setup();

        expect(screen.getByText('None Selected')).toBeInTheDocument();
        expect(screen.getByText('No information to display.')).toBeInTheDocument();
    });

    it('renders the node in the entity panel when member is clicked', async () => {
        const screen = await setup();
        const member = assetGroupMembers.members[0];
        const listItem = screen.getByText(member.name);
        const entityPanel = screen.getByTestId('explore_entity-information-panel');

        fireEvent.click(listItem);
        const header = await waitFor(() => screen.getByText('Object Information'));

        expect(header).toBeInTheDocument();
        expect(entityPanel).toHaveTextContent(member.name);
    });

    it('renders a link to the explore page when member is clicked', async () => {
        const screen = await setup();
        const member = assetGroupMembers.members[0];
        const listItem = screen.getByText(member.name);

        fireEvent.click(listItem);
        const link = screen.getByTestId('group-management_explore-link');
        expect(link).toBeInTheDocument();
    });
});
