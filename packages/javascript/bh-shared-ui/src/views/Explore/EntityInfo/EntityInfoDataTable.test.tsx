import { createMemoryHistory } from 'history';
import { RequestHandler, rest } from 'msw';
import { setupServer } from 'msw/node';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../../graphSchema';
import { act, render, screen } from '../../../test-utils';
import { allSections } from '../../../utils';
import EntityInfoDataTable from './EntityInfoDataTable';

const objectId = 'fake-object-id';
const adGpoSections = allSections[ActiveDirectoryNodeKind.GPO]!(objectId);
const azKeyVaultSections = allSections[AzureNodeKind.KeyVault]!(objectId);

const queryCount = {
    controllers: {
        count: 0,
        limit: 128,
        skip: 0,
        data: [],
    },
    ous: {
        count: 8,
        limit: 128,
        skip: 0,
        data: [],
    },
    computers: {
        count: 3003,
        limit: 128,
        skip: 0,
        data: [],
    },
    users: {
        count: 1998,
        limit: 128,
        skip: 0,
        data: [],
    },
    ['tier-zero']: {
        count: 55,
        limit: 128,
        skip: 0,
        data: [],
    },
} as const;

const keyVaultTest = {
    KeyReaders: {
        count: 0,
        limit: 128,
        skip: 0,
        data: [],
    },
    CertificateReaders: {
        count: 8,
        limit: 128,
        skip: 0,
        data: [],
    },
    SecretReaders: {
        count: 3003,
        limit: 128,
        skip: 0,
        data: [],
    },
    AllReaders: {
        count: 1998,
        limit: 128,
        skip: 0,
        data: [],
    },
} as const;

const handlers: Array<RequestHandler> = [
    rest.get(`api/v2/gpos/${objectId}/:asset`, (req, res, ctx) => {
        const asset = req.params.asset as keyof typeof queryCount;
        return res(ctx.json(queryCount[asset]));
    }),

    rest.get(`api/v2/azure/key-vaults*`, (req, res, ctx) => {
        if (req.url.searchParams.get('related_entity_type') === 'all-readers') {
            return res(ctx.json(keyVaultTest.AllReaders));
        }

        if (req.url.searchParams.get('related_entity_type') === 'key-readers') {
            return res(ctx.json(keyVaultTest.KeyReaders));
        }

        if (req.url.searchParams.get('related_entity_type') === 'secret-readers') {
            return res(ctx.json(keyVaultTest.SecretReaders));
        }

        if (req.url.searchParams.get('related_entity_type') === 'certificate-readers') {
            return res(ctx.json(keyVaultTest.CertificateReaders));
        }
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoDataTable', () => {
    describe('Node count for nested table that counts all sections', () => {
        it('sums nested section node counts', async () => {
            render(<EntityInfoDataTable {...adGpoSections[0]} />);
            const sum = await screen.findByText('5,064');
            expect(sum).not.toBeNull();
        });

        it('displays ! icon when one of the Affected Object calls fail', async () => {
            console.error = vi.fn();
            server.use(
                rest.get(`api/v2/gpos/${objectId}/ous`, (req, res, ctx) => {
                    return res(ctx.status(500));
                })
            );

            render(<EntityInfoDataTable {...adGpoSections[0]} />);

            const errorIcon = await screen.findByTestId('ErrorOutlineIcon');

            expect(errorIcon).not.toBeNull();
        });

        it('displays 0 when a given sections returns empty, and sums the rest of the sections correctly', async () => {
            const url = `?expandedPanelSections=${adGpoSections[0].label}`;
            const history = createMemoryHistory({ initialEntries: [url] });

            server.use(
                rest.get(`api/v2/gpos/${objectId}/ous`, (req, res, ctx) => {
                    const _ous = { ...queryCount.ous, count: undefined };
                    console.log(_ous);
                    return res(ctx.json(_ous));
                })
            );

            render(<EntityInfoDataTable {...adGpoSections[0]} />, { history });

            const sum = await screen.findAllByText('5,056');
            expect(sum).not.toBeNull();

            const zero = await screen.findByText('0');
            expect(zero.textContent).toBe('0');
        });
    });

    describe('Node count for Vault Readers nested table', () => {
        it('Verify Vault Reader count is the count returned by All Readers', async () => {
            const url = `?expandedPanelSections=${azKeyVaultSections[0].label}`;
            const history = createMemoryHistory({ initialEntries: [url] });

            render(<EntityInfoDataTable {...azKeyVaultSections[0]} />, { history });

            // verify the vault reader count is as expected
            const vaultReadersHeader = await screen.findByText('Vault Readers');
            await act(() => expect(vaultReadersHeader).toBeInTheDocument()); // verify accordion is open and showing all options
            const vaultReadersCount = vaultReadersHeader.nextElementSibling;
            expect(vaultReadersCount).toHaveTextContent('1,998');

            // verify the key readers count is as expected
            const keyReadersHeader = await screen.findByText('Key Readers');
            const keyReadersCount = keyReadersHeader.nextElementSibling;
            expect(keyReadersCount).toHaveTextContent('0');

            // verify the certificate readers count is as expected
            const certReadersHeader = await screen.findByText('Certificate Readers');
            const certReadersCount = certReadersHeader.nextElementSibling;
            expect(certReadersCount).toHaveTextContent('8');

            // verify the secret readers count is as expected
            const secretReadersHeader = await screen.findByText('Secret Readers');
            const secretReadersCount = secretReadersHeader.nextElementSibling;
            expect(secretReadersCount).toHaveTextContent('3,003');

            // verify the all readers count is as expected
            const allReadersHeader = await screen.findByText('All Readers');
            const allReadersCount = allReadersHeader.nextElementSibling;
            expect(allReadersCount).toHaveTextContent('1,998');
        });
    });
});
