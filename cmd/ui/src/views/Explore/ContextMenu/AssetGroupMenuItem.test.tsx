import { act } from 'react-dom/test-utils';
import { render, screen } from 'src/test-utils';
import userEvent from '@testing-library/user-event';
import { setupServer } from 'msw/node';
import { rest } from 'msw';
import AssetGroupMenuItem from './AssetGroupMenuItem';
import { apiClient } from 'bh-shared-ui';

describe('ContextMenu', async () => {
    const tierZeroAssetGroup = { id: '1', name: 'high value' };
    const ownedAssetGroup = { id: '2', name: 'owned' };

    const server = setupServer(
        rest.get('/api/v2/asset-groups/:assetGroupId/members', (req, res, ctx) => {
            // handle `tier zero` requests
            if (req.params.assetGroupId === tierZeroAssetGroup.id) {
                return res(
                    ctx.json({
                        data: {
                            members: [],
                        },
                    })
                );
            } else if (req.params.assetGroupId === ownedAssetGroup.id) {
                // handle `owned` requests
                return res(
                    ctx.json({
                        data: {
                            members: [{ custom_member: true }],
                        },
                    })
                );
            } else {
                return res(ctx.json({}));
            }
        }),
        rest.post('/api/v2/asset-groups/:assetGroupId/selectors', (req, res, ctx) => {
            return res(ctx.json({}));
        })
    );

    beforeAll(() => server.listen());

    afterEach(() => {
        server.resetHandlers();
    });

    afterAll(() => server.close());

    it('handles adding to an asset group', async () => {
        await act(async () => {
            render(
                <AssetGroupMenuItem assetGroupId={tierZeroAssetGroup.id} assetGroupName={tierZeroAssetGroup.name} />,
                {
                    initialState: {
                        entityinfo: {
                            selectedNode: {
                                name: 'foo',
                                id: '1234',
                                type: 'User',
                            },
                        },
                        assetgroups: {
                            assetGroups: [{ tag: 'admin_tier_0', id: tierZeroAssetGroup.id }],
                        },
                    },
                }
            );
        });

        const user = userEvent.setup();
        const addToHighValueSpy = vi.spyOn(apiClient, 'updateAssetGroupSelector');

        const addToHighValueButton = screen.getByRole('menuitem', { name: /add to high value/i });
        expect(addToHighValueButton).toBeInTheDocument();

        await user.click(addToHighValueButton);
        expect(addToHighValueSpy).toHaveBeenCalledTimes(1);
        expect(addToHighValueSpy).toHaveBeenCalledWith(tierZeroAssetGroup.id, [
            {
                action: 'add',
                selector_name: '1234',
                sid: '1234',
            },
        ]);
    });

    it('handles removing from an asset group', async () => {
        await act(async () => {
            await render(
                <AssetGroupMenuItem assetGroupId={ownedAssetGroup.id} assetGroupName={ownedAssetGroup.name} />,
                {
                    initialState: {
                        entityinfo: {
                            selectedNode: {
                                name: 'foo',
                                id: '1234',
                                type: 'User',
                            },
                        },
                        assetgroups: {
                            assetGroups: [{ tag: 'owned', id: ownedAssetGroup.id }],
                        },
                    },
                }
            );
        });

        const user = userEvent.setup();
        const removeFromOwnedSpy = vi.spyOn(apiClient, 'updateAssetGroupSelector');

        const removeFromOwnedButton = screen.getByRole('menuitem', { name: /remove from owned/i });
        expect(removeFromOwnedButton).toBeInTheDocument();

        await user.click(removeFromOwnedButton);
        expect(removeFromOwnedSpy).toHaveBeenCalledTimes(1);
        expect(removeFromOwnedSpy).toHaveBeenCalledWith(ownedAssetGroup.id, [
            {
                action: 'remove',
                selector_name: '1234',
                sid: '1234',
            },
        ]);
    });

    it('renders null if network fails to return valid asset group membership list', async () => {
        render(<AssetGroupMenuItem assetGroupId={'3'} assetGroupName={'blah'} />, {});

        expect(document.body.firstChild).toBeEmptyDOMElement();
    });
});
