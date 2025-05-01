import userEvent from '@testing-library/user-event';
import { createMemoryHistory } from 'history';
import { setupServer } from 'msw/node';
import { Route, Routes } from 'react-router-dom';
import { tierHandlers } from '../../../mocks';
import { render, screen } from '../../../test-utils';
import { apiClient } from '../../../utils';
import { MembersList } from './MembersList';

const handlers = [...tierHandlers];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const membersListSpy = vi.spyOn(apiClient, 'getAssetGroupSelectorMembers');

describe('MembersList', () => {
    it('sorting the list updates the list by changing the call made to the API', async () => {
        const user = userEvent.setup();

        const history = createMemoryHistory({
            initialEntries: ['/tier-management/details/tag/1/selector/1'],
        });

        render(
            <Routes>
                <Route path={'/'} element={<MembersList selected='1' onClick={vi.fn()} />} />;
                <Route
                    path={'/tier-management/details/tag/:tagId/selector/:selectorId'}
                    element={<MembersList selected='1' onClick={vi.fn()} />}
                />
            </Routes>,
            { history }
        );

        expect(membersListSpy).toBeCalledWith('1', '1', 0, 128, 'name');

        await user.click(screen.getByText('Objects', { exact: false }));

        expect(membersListSpy).toBeCalledWith('1', '1', 0, 128, '-name');
    });
});
