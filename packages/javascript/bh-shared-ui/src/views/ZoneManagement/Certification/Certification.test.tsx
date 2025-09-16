import userEvent from '@testing-library/user-event';
import { CertificationManual, CertificationRevoked } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from '../../../test-utils';
import { apiClient } from '../../../utils';
import Certification from './Certification';

//TODO -- add selection to test once it's no longer mocked in the component
const certifyMembersSpy = vi.spyOn(apiClient, 'updateAssetGroupTagCertification');

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
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Certification', () => {
    it('submits the selected items for certification with a note', async () => {
        const { container } = render(<Certification></Certification>);
        const textNote = 'a note';
        const certifyButton = await screen.findByText('Certify');
        expect(certifyButton).toBeInTheDocument();

        await user.click(certifyButton);
        const noteDialog = await screen.findByRole('dialog');
        expect(noteDialog).toBeInTheDocument();

        const input = container.querySelector('#textNote');
        await user.type(input!, textNote);
        expect(input).toHaveValue(textNote);
        const saveNote = await screen.findByText('Save Note');
        await user.click(saveNote);

        expect(certifyMembersSpy).toHaveBeenCalledTimes(1);

        expect(certifyMembersSpy).toHaveBeenCalledWith({
            member_ids: [777, 290, 91],
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
        render(<Certification></Certification>);
        const certifyButton = await screen.findByText('Certify');
        expect(certifyButton).toBeInTheDocument();

        await user.click(certifyButton);
        const noteDialog = await screen.findByRole('dialog');
        expect(noteDialog).toBeInTheDocument();

        const skipNote = await screen.findByText('Skip Note');
        await user.click(skipNote);

        expect(certifyMembersSpy).toHaveBeenCalledTimes(1);
        expect(certifyMembersSpy).toHaveBeenCalledWith({
            member_ids: [777, 290, 91],
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
        render(<Certification></Certification>);
        const revokeButton = await screen.findByText('Revoke');
        expect(revokeButton).toBeInTheDocument();

        await user.click(revokeButton);
        const noteDialog = await screen.findByRole('dialog');
        expect(noteDialog).toBeInTheDocument();

        const skipNote = await screen.findByText('Skip Note');
        await user.click(skipNote);

        expect(certifyMembersSpy).toHaveBeenCalledTimes(1);
        expect(certifyMembersSpy).toHaveBeenCalledWith({
            member_ids: [777, 290, 91],
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
        // TODO once no longer mocked
    });
    it('displays an error notification if the certification was unsuccessful', async () => {
        server.use(
            rest.post(`/api/v2/asset-group-tags/certifications`, async (_, res, ctx) => {
                return res(ctx.status(500));
            })
        );

        render(<Certification></Certification>);
        const certifyButton = await screen.findByText('Certify');
        expect(certifyButton).toBeInTheDocument();

        await user.click(certifyButton);
        const noteDialog = await screen.findByRole('dialog');
        expect(noteDialog).toBeInTheDocument();

        const skipNote = await screen.findByText('Skip Note');
        await user.click(skipNote);

        expect(certifyMembersSpy).toHaveBeenCalledTimes(1);
        expect(certifyMembersSpy).toHaveBeenCalledWith({
            member_ids: [777, 290, 91],
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
