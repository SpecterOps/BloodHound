import { rest } from 'msw';

export const defaultConfigurationResponse = { data: [] };

export const mockGetConfigurationHandler = (response = defaultConfigurationResponse) => {
    return rest.get('/api/v2/config', (_req, res, ctx) => res(ctx.json(response)));
};
