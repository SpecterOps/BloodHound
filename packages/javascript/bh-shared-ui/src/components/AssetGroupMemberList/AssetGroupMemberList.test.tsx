import { rest } from 'msw';
import { setupServer } from 'msw/node';
import AssetGroupMemberList from './AssetGroupMemberList';
import { fireEvent, render, waitFor } from '../../test-utils';
import { createAssetGroup, createAssetGroupMembers } from '../../mocks/factories';

const assetGroup = createAssetGroup();
const assetGroupMembers = createAssetGroupMembers();

const server = setupServer(
    rest.get('/api/v2/asset-groups/1/members', (req, res, ctx) => {
        return res(
            ctx.json({
                count: assetGroupMembers.members.length,
                limit: 100,
                skip: 0,
                data: assetGroupMembers,
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('AssetGroupMemberList', () => {
    const setup = () => {
        const handleSelectMember = vi.fn();
        const screen = render(
            <AssetGroupMemberList assetGroup={assetGroup} filter={{}} onSelectMember={handleSelectMember} />
        );
        return { screen, handleSelectMember };
    };

    it('Should display headers for member name and count', () => {
        const { screen } = setup();
        expect(screen.getByText('Name')).toBeInTheDocument();
        expect(screen.getByText('Custom Member')).toBeInTheDocument();
    });

    it('Should display a list of the asset group members', () => {
        const { screen } = setup();
        waitFor(() => {
            for (const member of assetGroupMembers.members) {
                expect(screen.getByText(member.name)).toBeInTheDocument();
            }
        });
    });

    it('Should call handler when a member is clicked', async () => {
        const { screen, handleSelectMember } = setup();
        const testMember = assetGroupMembers.members[0];
        const entry = await waitFor(() => screen.getByText(testMember.name));
        fireEvent.click(entry);
        expect(handleSelectMember).toHaveBeenCalledWith(testMember);
    });
});
