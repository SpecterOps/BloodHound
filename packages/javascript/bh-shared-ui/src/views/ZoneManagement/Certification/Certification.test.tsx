import { faker } from '@faker-js/faker';
import userEvent from '@testing-library/user-event';
import { AssetGroupTagsCertification, CertificationManual, CertificationRevoked } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import * as reactQuery from 'react-query';
import { render, screen } from '../../../test-utils';
import { apiClient } from '../../../utils';
import Certification from './Certification';

const certifyMembersSpy = vi.spyOn(apiClient, 'updateAssetGroupTagCertification');
const useInfiniteQuerySpy = vi.spyOn(reactQuery, 'useInfiniteQuery');

const mockCertificationData: AssetGroupTagsCertification = {
    data: {
        members: [
            {
                id: 1,
                object_id: faker.datatype.uuid(),
                environment_id: '00000-00000-00001',
                primary_kind: 'User',
                name: 'ADMIN@WRAITH.CORP',
                created_at: '2025-04-24T20:28:45.676055Z',
                asset_group_tag_id: 111,
                certified_by: '',
                certified: 0,
            },
            {
                id: 2,
                object_id: faker.datatype.uuid(),
                environment_id: '00000-00000-00001',
                primary_kind: 'User',
                name: 'ADMIN@PHANTOM.CORP',
                created_at: '2025-04-24T20:28:45.676055Z',
                asset_group_tag_id: 111,
                certified_by: '',
                certified: 0,
            },
            {
                id: 3,
                object_id: faker.datatype.uuid(),
                environment_id: '00000-00000-00001',
                primary_kind: 'User',
                name: 'ADMIN@GHOST.CORP',
                created_at: '2025-04-24T20:28:45.676055Z',
                asset_group_tag_id: 111,
                certified_by: '',
                certified: 0,
            },
        ],
    },
    count: 3,
    limit: 15,
    skip: 0,
};

const addNotificationMock = vi.fn();

vi.mock('../../../providers', async () => {
    const actual = await vi.importActual('../../../providers');
    return {
        ...actual,
        useNotifications: () => {
            return { addNotification: addNotificationMock };
        },
    };
});

const server = setupServer(
    rest.post(`/api/v2/asset-group-tags/certifications`, async (_, res, ctx) => {
        return res(ctx.status(200));
    })
);

const user = userEvent.setup();

beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
});
afterAll(() => server.close());

describe('Certification', () => {
    it('submits the selected items for certification with a note', async () => {
        //@ts-ignore
        useInfiniteQuerySpy.mockReturnValue({
            data: {
                pages: [mockCertificationData],
                pageParams: [],
            },
            fetchNextPage: vi.fn(),
            isLoading: false,
            isFetching: false,
            isSuccess: true,
            isError: false,
            refetch: vi.fn(),
        });
        const { container } = render(<Certification></Certification>);
        const selectAllCheckbox = await screen.findByTestId('certification-table-select-all');
        expect(selectAllCheckbox).toBeInTheDocument();
        await user.click(selectAllCheckbox);

        const certifyButton = await screen.findByText('Certify');
        expect(certifyButton).toBeInTheDocument();

        await user.click(certifyButton);
        const noteDialog = await screen.findByRole('dialog');
        expect(noteDialog).toBeInTheDocument();

        const input = container.querySelector('#textNote');
        const textNote = 'a note';
        await user.type(input!, textNote);
        expect(input).toHaveValue(textNote);
        const saveNote = await screen.findByText('Save Note');
        await user.click(saveNote);

        expect(certifyMembersSpy).toHaveBeenCalledTimes(1);

        expect(certifyMembersSpy).toHaveBeenCalledWith({
            member_ids: [1, 2, 3],
            action: CertificationManual,
            note: textNote,
        });
        expect(addNotificationMock).toBeCalledWith(
            'Selected Certification Successful',
            'zone-management_update-certification_success',
            {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            }
        );
    });
    it('submits the selected items for certification without a note', async () => {
        //@ts-ignore
        useInfiniteQuerySpy.mockReturnValue({
            data: {
                pages: [mockCertificationData],
                pageParams: [],
            },
            fetchNextPage: vi.fn(),
            isLoading: false,
            isFetching: false,
            isSuccess: true,
            isError: false,
            refetch: vi.fn(),
        });
        render(<Certification></Certification>);

        const selectAllCheckbox = await screen.findByTestId('certification-table-select-all');
        expect(selectAllCheckbox).toBeInTheDocument();
        await user.click(selectAllCheckbox);

        const certifyButton = await screen.findByText('Certify');
        expect(certifyButton).toBeInTheDocument();

        await user.click(certifyButton);
        const noteDialog = await screen.findByRole('dialog');
        expect(noteDialog).toBeInTheDocument();

        const skipNote = await screen.findByText('Skip Note');
        await user.click(skipNote);

        expect(certifyMembersSpy).toHaveBeenCalledTimes(1);
        expect(certifyMembersSpy).toHaveBeenCalledWith({
            member_ids: [1, 2, 3],
            action: CertificationManual,
        });
        expect(addNotificationMock).toBeCalledWith(
            'Selected Certification Successful',
            'zone-management_update-certification_success',
            {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            }
        );
    });
    it('submits the selected items for revocation', async () => {
        //@ts-ignore
        useInfiniteQuerySpy.mockReturnValue({
            data: {
                pages: [mockCertificationData],
                pageParams: [],
            },
            fetchNextPage: vi.fn(),
            isLoading: false,
            isFetching: false,
            isSuccess: true,
            isError: false,
            refetch: vi.fn(),
        });

        render(<Certification></Certification>);

        const selectAllCheckbox = await screen.findByTestId('certification-table-select-all');
        expect(selectAllCheckbox).toBeInTheDocument();
        await user.click(selectAllCheckbox);

        const revokeButton = await screen.findByText('Revoke');
        expect(revokeButton).toBeInTheDocument();

        await user.click(revokeButton);
        const noteDialog = await screen.findByRole('dialog');
        expect(noteDialog).toBeInTheDocument();

        const skipNote = await screen.findByText('Skip Note');
        await user.click(skipNote);

        expect(certifyMembersSpy).toHaveBeenCalledTimes(1);
        expect(certifyMembersSpy).toHaveBeenCalledWith({
            member_ids: [1, 2, 3],
            action: CertificationRevoked,
        });
        expect(addNotificationMock).toBeCalledWith(
            'Selected Revocation Successful',
            'zone-management_update-certification_success',
            {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            }
        );
    });
    it('does not call the API if no items are selected', async () => {
        //@ts-ignore
        useInfiniteQuerySpy.mockReturnValue({
            data: {
                pages: [mockCertificationData],
                pageParams: [],
            },
            fetchNextPage: vi.fn(),
            isLoading: false,
            isFetching: false,
            isSuccess: true,
            isError: false,
            refetch: vi.fn(),
        });
        const { container } = render(<Certification></Certification>);

        const certifyButton = await screen.findByText('Certify');
        expect(certifyButton).toBeInTheDocument();

        await user.click(certifyButton);
        const noteDialog = await screen.findByRole('dialog');
        expect(noteDialog).toBeInTheDocument();

        const input = container.querySelector('#textNote');
        const textNote = 'a note';
        await user.type(input!, textNote);
        expect(input).toHaveValue(textNote);
        const saveNote = await screen.findByText('Save Note');
        await user.click(saveNote);

        expect(certifyMembersSpy).toHaveBeenCalledTimes(0);

        expect(addNotificationMock).toBeCalledWith(
            'Members must be selected for certification',
            'zone-management_update-certification_no-members',
            {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            }
        );
    });
    it('displays an error notification if the certification was unsuccessful', async () => {
        server.use(
            rest.post(`/api/v2/asset-group-tags/certifications`, async (_, res, ctx) => {
                return res(ctx.status(500));
            })
        );

        //@ts-ignore
        useInfiniteQuerySpy.mockReturnValue({
            data: {
                pages: [mockCertificationData],
                pageParams: [],
            },
            fetchNextPage: vi.fn(),
            isLoading: false,
            isFetching: false,
            isSuccess: true,
            isError: false,
            refetch: vi.fn(),
        });

        render(<Certification></Certification>);

        const selectAllCheckbox = await screen.findByTestId('certification-table-select-all');
        expect(selectAllCheckbox).toBeInTheDocument();
        await user.click(selectAllCheckbox);

        const certifyButton = await screen.findByText('Certify');
        expect(certifyButton).toBeInTheDocument();

        await user.click(certifyButton);
        const noteDialog = await screen.findByRole('dialog');
        expect(noteDialog).toBeInTheDocument();

        const skipNote = await screen.findByText('Skip Note');
        await user.click(skipNote);

        expect(certifyMembersSpy).toHaveBeenCalledTimes(1);
        expect(certifyMembersSpy).toHaveBeenCalledWith({
            member_ids: [1, 2, 3],
            action: CertificationManual,
        });
        expect(addNotificationMock).toBeCalledWith(
            'There was an error updating certification',
            `zone-management_update-certification_error`,
            {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            }
        );
    });
});
