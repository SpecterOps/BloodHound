import { rest } from 'msw';
import { createFeatureFlags } from '../factories/featureFlags';

const handlers = [
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.status(200),
            ctx.json({
                data: createFeatureFlags([
                    {
                        key: 'deep-linking',
                        enabled: true,
                    },
                ]),
            })
        );
    }),
];

export default handlers;
