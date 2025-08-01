import { type ScheduledJobDisplay } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';

import { renderHook, waitFor } from '../test-utils';
import {
    FETCH_ERROR_KEY,
    FETCH_ERROR_MESSAGE,
    NO_PERMISSION_KEY,
    NO_PERMISSION_MESSAGE,
    PERSIST_NOTIFICATION,
    toCollected,
    toFormatted,
    toMins,
    useFinishedJobsQuery,
} from './finishedJobs';

const addNotificationMock = vi.fn();
const dismissNotificationMock = vi.fn();
const checkPermissionMock = vi.fn();

vi.mock('../../providers', async () => {
    const actual = await vi.importActual('../../providers');
    return {
        ...actual,
        useNotifications: () => ({
            addNotification: addNotificationMock,
            dismissNotification: dismissNotificationMock,
        }),
    };
});

vi.mock('../../hooks', async () => {
    const actual = await vi.importActual('../../hooks');
    return {
        ...actual,
        usePermissions: () => ({
            checkPermission: checkPermissionMock,
        }),
    };
});

const MOCK_FINISHED_JOB: ScheduledJobDisplay = {
    id: 22,
    client_id: '718c9b04-9394-42c0-9d53-c87b689e2d92',
    client_name: 'GOAD',
    event_id: 123,
    execution_time: '2024-08-15T21:24:52.366579Z',
    start_time: '2024-08-15T21:25:21.990437Z',
    end_time: '2024-08-15T21:26:43.033448Z',
    status: 2,
    status_message: 'The service collected successfully',
    session_collection: true,
    local_group_collection: true,
    ad_structure_collection: true,
    cert_services_collection: true,
    ca_registry_collection: true,
    dc_registry_collection: true,
    all_trusted_domains: true,
    domain_controller: '',
    ous: [],
    domains: [],
    domain_results: [],
};

export const MOCK_FINISHED_JOBS_RESPONSE = {
    count: 20,
    data: new Array(10).fill(MOCK_FINISHED_JOB).map((item, index) => ({
        ...item,
        id: index,
        status: (index % 10) - 1,
    })),
    limit: 10,
    skip: 10,
};

const server = setupServer(
    rest.get('/api/v2/jobs/finished', (req, res, ctx) => res(ctx.json(MOCK_FINISHED_JOBS_RESPONSE)))
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('toCollected', () => {
    it('shows the collection methods for the given job', () => {
        expect(toCollected(MOCK_FINISHED_JOB)).toBe(
            'Sessions, Local Groups, AD Structure, Certificate Services, CA Registry, DC Registry, All Trusted Domains'
        );
    });

    it('shows some collection methods for the given job', () => {
        const NO_COLLECTIONS_JOB = {
            ...MOCK_FINISHED_JOB,
            session_collection: false,
            local_group_collection: false,
            ad_structure_collection: false,
            cert_services_collection: false,
            ca_registry_collection: false,
            dc_registry_collection: false,
            all_trusted_domains: false,
        };
        expect(toCollected(NO_COLLECTIONS_JOB)).toBe('');
    });

    it('shows no collection methods for the given job', () => {
        const SOME_COLLECTIONS_JOB = {
            ...MOCK_FINISHED_JOB,
            session_collection: false,
            local_group_collection: false,
            ad_structure_collection: false,
            cert_services_collection: false,
        };
        expect(toCollected(SOME_COLLECTIONS_JOB)).toBe('CA Registry, DC Registry, All Trusted Domains');
    });
});

describe('toFormatted', () => {
    it('formats the date string', () => {
        const result = toFormatted('2024-01-01T15:30:00.500Z');
        // Server TZ might not match local dev TZ
        // Match format like '2024-01-01 09:30 CST'
        expect(result).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2} [A-Z]{3,4}$/);
    });
});

describe('toMins', () => {
    it('shows an interval in mins', () => {
        expect(toMins('2024-01-01T15:30:00.500Z', '2024-01-02T03:00:00.000Z')).toBe('689 Min');
    });
});

describe('useFinishedJobsQuery', () => {
    afterEach(() => {
        addNotificationMock.mockReset();
        checkPermissionMock.mockReset();
        dismissNotificationMock.mockReset();
    });

    it('requests finished jobs', async () => {
        checkPermissionMock.mockImplementation(() => true);
        const { result } = renderHook(() => useFinishedJobsQuery({ page: 0, rowsPerPage: 10 }));
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(result.current.data.data.length).toBe(10);
    });

    it('shows "no permission" notification if lacking permission', async () => {
        checkPermissionMock.mockImplementation(() => false);
        const { result } = renderHook(() => useFinishedJobsQuery({ page: 0, rowsPerPage: 10 }));
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(addNotificationMock).toHaveBeenCalledWith(
            NO_PERMISSION_MESSAGE,
            NO_PERMISSION_KEY,
            PERSIST_NOTIFICATION
        );
    });

    it('does not request finished jobs if lacking permission', async () => {
        checkPermissionMock.mockImplementation(() => false);
        const { result } = renderHook(() => useFinishedJobsQuery({ page: 0, rowsPerPage: 10 }));
        await waitFor(() => expect(result.current.isLoading).toBe(false));
        expect(result.current.data).toBeUndefined();
    });

    it('shows an error notification if there is an error fetching', async () => {
        server.use(rest.get('/api/v2/jobs/finished', (req, res, ctx) => res(ctx.status(400))));
        checkPermissionMock.mockImplementation(() => true);
        const { result } = renderHook(() => useFinishedJobsQuery({ page: 0, rowsPerPage: 10 }));
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(addNotificationMock).toHaveBeenCalledWith(FETCH_ERROR_MESSAGE, FETCH_ERROR_KEY);
    });
});
