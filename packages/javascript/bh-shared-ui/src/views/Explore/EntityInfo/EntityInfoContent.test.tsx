import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../../graphSchema';
import { render, screen, waitForElementToBeRemoved } from '../../../test-utils';
import { EntityKinds } from '../../../utils';
import { ObjectInfoPanelContextProvider } from '../providers/ObjectInfoPanelProvider';
import EntityInfoContent from './EntityInfoContent';

const server = setupServer(
    rest.get('/api/v2/azure/roles', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    kind: 'AZRole',
                    props: {},
                    active_assignments: 0,
                    pim_assignments: 0,
                },
            })
        );
    }),
    rest.get('/api/v2/base/*', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    kind: ActiveDirectoryNodeKind.Entity,
                    props: { objectid: 'test' },
                },
            })
        );
    }),
    rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    nodes: {
                        '42': {
                            kind: 'Unknown',
                            properties: { objectid: 'unknown kind' },
                        },
                    },
                },
            })
        );
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    })
);

const EntityInfoContentWithProvider = ({
    testId,
    nodeType,
    databaseId,
}: {
    testId: string;
    nodeType: EntityKinds | string;
    databaseId?: string;
}) => (
    <ObjectInfoPanelContextProvider>
        <EntityInfoContent id={testId} nodeType={nodeType} databaseId={databaseId} />
    </ObjectInfoPanelContextProvider>
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoContent', () => {
    it('AZRole information panel will not display a section for PIM Assignments', async () => {
        const testId = '1';
        const nodeType = AzureNodeKind.Role;

        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} />);
        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));
        expect(screen.queryByText('PIM Assignments')).not.toBeInTheDocument();
    });
});

describe('EntityObjectInformation', () => {
    it('Calls the `Base` endpoint for a LocalGroup type node', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.LocalGroup;

        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} />);

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('test')).toBeInTheDocument();
    });

    it('Calls the `Base` endpoint for a LocalUser type node', async () => {
        const testId = '1';
        const nodeType = ActiveDirectoryNodeKind.LocalUser;

        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} />);

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('test')).toBeInTheDocument();
    });

    it('Calls the cypher search endpoint for a node with a type that is not in our schema', async () => {
        const testId = '1';
        const nodeType = 'Unknown';
        const databaseId = '42';

        render(<EntityInfoContentWithProvider testId={testId} nodeType={nodeType} databaseId={databaseId} />);

        await waitForElementToBeRemoved(() => screen.getByTestId('entity-object-information-skeleton'));

        expect(await screen.findByText('unknown kind')).toBeInTheDocument();
    });
});
