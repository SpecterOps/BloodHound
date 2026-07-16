import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { renderHook, waitFor } from '../test-utils';
import { apiClient } from '../utils';
import { useGetUser } from './useBloodHoundUsers';

const MOCK_USER_ID = '718c9b04-9394-42c0-9d53-c87b689e2d92';

const MOCK_USER = {
    id: MOCK_USER_ID,
    first_name: 'Ada',
    last_name: 'Lovelace',
    email_address: 'ada@example.com',
};

const server = setupServer(
    rest.get(`/api/v2/bloodhound-users/${MOCK_USER_ID}`, (req, res, ctx) => {
        return res(ctx.json({ data: MOCK_USER }));
    })
);

beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
    vi.restoreAllMocks();
});
afterAll(() => server.close());

describe('useGetUser', () => {
    it('does not request a user when no userId is provided', async () => {
        const getUserSpy = vi.spyOn(apiClient, 'getUser');
        const { result } = renderHook(() => useGetUser(undefined));

        expect(getUserSpy).not.toHaveBeenCalled();
        expect(result.current.isLoading).toBe(false);
        expect(result.current.data).toBeUndefined();
    });

    it('requests the user when a userId is provided', async () => {
        const getUserSpy = vi.spyOn(apiClient, 'getUser');
        const { result } = renderHook(() => useGetUser(MOCK_USER_ID));

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(getUserSpy).toHaveBeenCalledWith(
            MOCK_USER_ID,
            expect.objectContaining({ signal: expect.any(AbortSignal) })
        );
        expect(result.current.data).toEqual(MOCK_USER);
    });
});
