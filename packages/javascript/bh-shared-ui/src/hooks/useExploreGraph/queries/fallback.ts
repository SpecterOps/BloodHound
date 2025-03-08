import { ExploreGraphQuery } from './utils';

export const fallbackQuery: ExploreGraphQuery = {
    getQueryConfig: () => ({ enabled: false }),
    getErrorMessage: () => ({ message: 'An unknown error occurred.', key: 'unknownGraphError' }),
};
