import { setupServer } from 'msw/node';
import { createMockAssetGroup, createMockAssetGroupMembers, createMockSearchResults } from '../../mocks/factories';
import { act, fireEvent, render, waitFor } from '../../test-utils';
import { AUTOCOMPLETE_PLACEHOLDER } from './AssetGroupAutocomplete';
import AssetGroupEdit from './AssetGroupEdit';
import { rest } from 'msw';

const assetGroup = createMockAssetGroup();
const assetGroupMembers = createMockAssetGroupMembers();
const searchResults = createMockSearchResults();

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
    }),
    rest.get('/api/v2/search', (req, res, ctx) => {
        return res(
            ctx.json({
                data: searchResults,
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('AssetGroupEdit', () => {
    const setup = async () => {
        return await act(async () => {
            return render(<AssetGroupEdit assetGroup={assetGroup} filter={{}} />);
        });
    };

    const escapeKey = {
        key: 'Escape',
        code: 'Escape',
        keyCode: 27,
        charCode: 27,
    };

    it('should display a searchbox with a placeholder when rendered', async () => {
        const screen = await setup();
        const input = screen.getByPlaceholderText(AUTOCOMPLETE_PLACEHOLDER);
        expect(input).toBeInTheDocument();
    });

    it('should display a total count of asset group members', async () => {
        const screen = await setup();
        const count = screen.getByText('Total Count').nextSibling.textContent;
        expect(count).toBe(assetGroupMembers.members.length.toString());
    });

    it('should display search results when the user enters text', async () => {
        const screen = await setup();
        const input = screen.getByRole('combobox');

        fireEvent.change(input, { target: { value: 'test' } });
        expect(input.value).toEqual('test');

        const result = await waitFor(() => screen.getByText('00001.TESTLAB.LOCAL'));
        expect(result).toBeInTheDocument();
    });

    it('should add an option and display the changelog when it is clicked', async () => {
        const screen = await setup();
        const selection = searchResults[0];

        const input = screen.getByRole('combobox');
        fireEvent.change(input, { target: { value: 'test' } });
        expect(input.value).toEqual('test');

        const result = await waitFor(() => screen.getByText(selection.name));
        fireEvent.click(result);
        fireEvent.keyDown(input, escapeKey);

        expect(screen.getByText(selection.name)).toBeInTheDocument();
        expect(screen.getByText(selection.objectid)).toBeInTheDocument();
        expect(screen.getByText('Cancel')).toBeInTheDocument();
        expect(screen.getByText('Confirm Changes')).toBeInTheDocument();
    });

    it('should remove the option from the changelog when the corresponding button is clicked', async () => {
        const screen = await setup();
        const selection = searchResults[0];

        const input = screen.getByRole('combobox');
        fireEvent.change(input, { target: { value: 'test' } });
        expect(input.value).toEqual('test');

        const result = await waitFor(() => screen.getByText(selection.name));
        fireEvent.click(result);
        fireEvent.keyDown(input, escapeKey);

        const removeButton = screen.getByText('xmark');
        fireEvent.click(removeButton);

        expect(screen.queryByText(selection.name)).toBeNull();
        expect(screen.queryByText(selection.objectid)).toBeNull();
    });
});
