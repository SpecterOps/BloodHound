import { ConfigurationKey } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from '../../test-utils';
import SalesMessage from './SalesMessage';

const configResponse = {
    data: [
        {
            key: ConfigurationKey.Tiering,
            value: { multi_tier_analysis_enabled: false, tier_limit: 3, label_limit: 10 },
        },
    ],
};

const server = setupServer();

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('SalesMessage', async () => {
    it('renders sales message when multi_tier_analysis_enabled is false', async () => {
        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configResponse));
            })
        );

        render(<SalesMessage />);

        expect(await screen.findByText(/Upgrade Privilege Zones/i)).toBeInTheDocument();
    });

    it('does not render sales message when multi_tier_analysis_enabled is true', () => {
        const configRes = {
            data: [
                {
                    key: ConfigurationKey.Tiering,
                    value: { multi_tier_analysis_enabled: true, tier_limit: 3, label_limit: 10 },
                },
            ],
        };

        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configRes));
            })
        );

        render(<SalesMessage />);

        expect(screen.queryByText(/Upgrade Privilege Zones/i)).not.toBeInTheDocument();
    });
});