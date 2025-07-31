import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, render, screen } from '../../test-utils';
import { FinishedJobsTable } from './FinishedJobsTable';
import { MOCK_FINISHED_JOBS_RESPONSE } from './finishedJobs.test';

const checkPermissionMock = vi.fn();

vi.mock('../../hooks', async () => {
    const actual = await vi.importActual('../../hooks');
    return {
        ...actual,
        usePermissions: () => ({
            checkPermission: checkPermissionMock,
        }),
    };
});

const server = setupServer(
    rest.get('/api/v2/jobs/finished', (req, res, ctx) => res(ctx.json(MOCK_FINISHED_JOBS_RESPONSE)))
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('FinishedJobsTable', () => {
    it('shows a loading state', () => {
        checkPermissionMock.mockImplementation(() => true);
        const { container } = render(<FinishedJobsTable />);

        // 1 loading skeleton for each column
        const children = container.querySelectorAll('.MuiSkeleton-pulse');
        expect(children.length).toBe(5);
    });

    it('shows a table with finished jobs', async () => {
        checkPermissionMock.mockImplementation(() => true);
        await act(async () => render(<FinishedJobsTable />));

        const jobStatus = await screen.findByText('Complete');
        expect(jobStatus).toHaveTextContent('Complete');
    });
});
