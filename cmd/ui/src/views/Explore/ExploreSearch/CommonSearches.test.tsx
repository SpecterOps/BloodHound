import { render, screen } from 'src/test-utils';
import CommonSearches, { getADSearches, getAZSearches } from './CommonSearches';
import userEvent from '@testing-library/user-event';
import { setupServer } from 'msw/node';
import { rest } from 'msw';

const server = setupServer(
    rest.get('/api/v2/queries', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        user_id: 'abcdefgh',
                        query: 'match (n) return n limit 5',
                        name: 'me save a query 1',
                    },
                    {
                        user_id: 'abcdefgh',
                        query: 'match (n) return n limit 5',
                        name: 'me save a query 2',
                    },
                ],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('CommonSearches', () => {
    const onClick = jest.fn();
    beforeEach(() => {
        onClick.mockReset();
        render(<CommonSearches onClickListItem={onClick} />);
    });

    it('renders headers', () => {
        const header = screen.getByText(/pre-built searches/i);
        const adTab = screen.getByRole('tab', { name: /active directory/i });
        const azTab = screen.getByRole('tab', { name: /azure/i });
        const userTab = screen.getByRole('tab', { name: /custom searches/i });

        expect(header).toBeInTheDocument();
        expect(adTab).toBeInTheDocument();
        expect(azTab).toBeInTheDocument();
        expect(userTab).toBeInTheDocument();

        expect(screen.getByRole('tab', { selected: true })).toHaveTextContent('Active Directory');
    });

    it('renders search list for the currently active tab', () => {
        const adSearches = getADSearches();
        const subheadersForAD = adSearches.map((element) => element.subheader);

        subheadersForAD.forEach((subheader) => {
            expect(screen.getByText(subheader)).toBeInTheDocument();
        });
    });

    it('renders a different list of queries when user switches tab', async () => {
        const user = userEvent.setup();

        // switch tabs to AZ
        const azureTab = screen.getByRole('tab', { name: /azure/i });
        await user.click(azureTab);

        const azSearches = getAZSearches();
        const subheadersForAZ = azSearches.map((element) => element.subheader);

        subheadersForAZ.forEach((subheader) => {
            expect(screen.getByText(subheader)).toBeInTheDocument();
        });
    });

    it(`fetches a user's saved queries when the 'custom searches' tab is clicked`, async () => {
        const user = userEvent.setup();

        // switch tabs to user searches
        const userTab = screen.getByRole('tab', { name: /custom searches/i });
        await user.click(userTab);

        const firstSavedQuery = screen.getAllByRole('button', { name: /me save a query/i });
        expect(firstSavedQuery).toHaveLength(2);
    });

    it('handles a click on each list item', async () => {
        const user = userEvent.setup();

        const adSearches = getADSearches();
        const { cypher, description } = adSearches[0].queries[0];

        const listItem = screen.getByRole('button', { name: description });
        expect(listItem).toBeInTheDocument();

        await user.click(listItem);

        expect(onClick).toHaveBeenCalledTimes(1);
        expect(onClick).toHaveBeenCalledWith(cypher);
    });
    // todo:
    it('handles the secondary action (delete) on each list item', async () => {});
});
