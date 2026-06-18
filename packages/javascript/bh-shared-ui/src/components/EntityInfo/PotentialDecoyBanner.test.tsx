// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import userEvent from '@testing-library/user-event';
import {
    AssetGroupTagTypeDecoy,
    SeedTypeObjectId,
    type AssetGroup,
    type AssetGroupTagSelector,
} from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { DECOY_OBJECT_TAG, TAG_DECOY_AGT } from '../../constants';
import { ActiveDirectoryKindProperties, ActiveDirectoryNodeKind, CommonKindProperties } from '../../graphSchema';
import { createAuthStateWithPermissions } from '../../mocks';
import { render, screen, waitFor } from '../../test-utils';
import { Permission } from '../../utils';
import PotentialDecoyBanner, { isPotentialDecoyUser } from './PotentialDecoyBanner';

const basePotentialDecoyProperties = {
    [ActiveDirectoryKindProperties.LastLogon]: 0,
    [ActiveDirectoryKindProperties.LastLogonTimestamp]: -1,
    [CommonKindProperties.Enabled]: true,
    [CommonKindProperties.Name]: 'svc-watch@TESTLAB.LOCAL',
    [CommonKindProperties.ObjectID]: 'S-1-5-21-570004220-2248230615-4072641716-5965',
    [CommonKindProperties.WhenCreated]: '2026-03-01T00:00:00Z',
};

const objectId = basePotentialDecoyProperties[CommonKindProperties.ObjectID];

const decoyTag = {
    id: 4,
    type: AssetGroupTagTypeDecoy,
    kind_id: 4,
    position: null,
    require_certify: null,
    name: 'Decoy',
    description: 'Decoy',
    analysis_enabled: null,
    glyph: null,
};

const decoyAssetGroup: AssetGroup = {
    id: 4,
    name: 'Decoy',
    tag: DECOY_OBJECT_TAG,
} as AssetGroup;

let featureFlagEnabled = true;
let graphWritePermission = true;
let selectorQueryParams: URLSearchParams | undefined;
let legacyMembersQueryParams: URLSearchParams | undefined;
let selectorRequestBody: any;
let deletedSelectorId: string | undefined;
let legacySelectorChangeset: any;
let agtSelectors: AssetGroupTagSelector[] = [];
let legacyMembers: Record<string, any>[] = [];

const server = setupServer(
    rest.get('/api/v2/features', (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [{ id: 1, key: 'tier_management_engine', enabled: featureFlagEnabled }],
            })
        );
    }),
    rest.get('/api/v2/self', (_req, res, ctx) => {
        return res(
            ctx.json({
                data: createAuthStateWithPermissions(graphWritePermission ? [Permission.GRAPH_DB_WRITE] : []).user,
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags', (_req, res, ctx) => {
        return res(ctx.json({ data: { tags: [decoyTag] } }));
    }),
    rest.get('/api/v2/asset-group-tags/:tagId/selectors', (req, res, ctx) => {
        selectorQueryParams = req.url.searchParams;
        return res(ctx.json({ data: { selectors: agtSelectors } }));
    }),
    rest.post('/api/v2/asset-group-tags/:tagId/selectors', async (req, res, ctx) => {
        selectorRequestBody = await req.json();
        return res(ctx.json({ data: { selector: { id: 1 } } }));
    }),
    rest.delete('/api/v2/asset-group-tags/:tagId/selectors/:selectorId', (req, res, ctx) => {
        deletedSelectorId = req.params.selectorId as string;
        return res(ctx.json({}));
    }),
    rest.get('/api/v2/asset-groups', (_req, res, ctx) => {
        return res(ctx.json({ data: { asset_groups: [decoyAssetGroup] } }));
    }),
    rest.get('/api/v2/asset-groups/:assetGroupId/members', (req, res, ctx) => {
        legacyMembersQueryParams = req.url.searchParams;
        return res(ctx.json({ data: { members: legacyMembers } }));
    }),
    rest.put('/api/v2/asset-groups/:assetGroupId/selectors', async (req, res, ctx) => {
        legacySelectorChangeset = await req.json();
        return res(ctx.json({}));
    })
);

beforeAll(() => server.listen());

beforeEach(() => {
    featureFlagEnabled = true;
    graphWritePermission = true;
    selectorQueryParams = undefined;
    legacyMembersQueryParams = undefined;
    selectorRequestBody = undefined;
    deletedSelectorId = undefined;
    legacySelectorChangeset = undefined;
    agtSelectors = [];
    legacyMembers = [];
});

afterEach(() => {
    server.resetHandlers();
    vi.restoreAllMocks();
});

afterAll(() => server.close());

describe('isPotentialDecoyUser', () => {
    beforeEach(() => {
        vi.useFakeTimers();
        vi.setSystemTime(new Date('2026-06-18T00:00:00Z'));
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('returns true for old enabled users with no recorded logon', () => {
        expect(isPotentialDecoyUser(ActiveDirectoryNodeKind.User, basePotentialDecoyProperties)).toBe(true);
    });

    it('returns false for non-user nodes', () => {
        expect(isPotentialDecoyUser(ActiveDirectoryNodeKind.Computer, basePotentialDecoyProperties)).toBe(false);
    });

    it('returns false for recently created users', () => {
        expect(
            isPotentialDecoyUser(ActiveDirectoryNodeKind.User, {
                ...basePotentialDecoyProperties,
                [CommonKindProperties.WhenCreated]: '2026-06-01T00:00:00Z',
            })
        ).toBe(false);
    });

    it('returns false for disabled users', () => {
        expect(
            isPotentialDecoyUser(ActiveDirectoryNodeKind.User, {
                ...basePotentialDecoyProperties,
                [CommonKindProperties.Enabled]: false,
            })
        ).toBe(false);
    });

    it('returns false for special accounts', () => {
        expect(
            isPotentialDecoyUser(ActiveDirectoryNodeKind.User, {
                ...basePotentialDecoyProperties,
                [CommonKindProperties.ObjectID]: 'S-1-5-21-570004220-2248230615-4072641716-500',
            })
        ).toBe(false);

        expect(
            isPotentialDecoyUser(ActiveDirectoryNodeKind.User, {
                ...basePotentialDecoyProperties,
                [CommonKindProperties.Name]: 'AZUREADSSOACC.TESTLAB.LOCAL',
            })
        ).toBe(false);
    });

    it('returns false for managed service accounts', () => {
        expect(
            isPotentialDecoyUser(ActiveDirectoryNodeKind.User, {
                ...basePotentialDecoyProperties,
                [ActiveDirectoryKindProperties.GMSA]: true,
            })
        ).toBe(false);

        expect(
            isPotentialDecoyUser(ActiveDirectoryNodeKind.User, {
                ...basePotentialDecoyProperties,
                [ActiveDirectoryKindProperties.MSA]: true,
            })
        ).toBe(false);
    });
});

describe('PotentialDecoyBanner', () => {
    const oldPotentialDecoyProperties = {
        ...basePotentialDecoyProperties,
        [CommonKindProperties.WhenCreated]: '2020-03-01T00:00:00Z',
    };

    it('filters AGT selectors by object ID and can mark a potential decoy', async () => {
        render(
            <PotentialDecoyBanner
                nodeType={ActiveDirectoryNodeKind.User}
                objectId={objectId}
                properties={oldPotentialDecoyProperties}
            />
        );

        expect(await screen.findByRole('alert')).toHaveTextContent(/might be a decoy/i);

        await waitFor(() => {
            expect(selectorQueryParams?.get('limit')).toBe('1');
            expect(selectorQueryParams?.get('type')).toBe(`eq:${SeedTypeObjectId}`);
            expect(selectorQueryParams?.get('value')).toBe(`eq:${objectId}`);
        });

        const markDecoySwitch = await screen.findByRole('checkbox', {
            name: /mark object as decoy/i,
        });

        await waitFor(() => {
            expect(markDecoySwitch).not.toBeDisabled();
        });

        await userEvent.click(markDecoySwitch);

        await waitFor(() => {
            expect(selectorRequestBody).toEqual({
                name: oldPotentialDecoyProperties[CommonKindProperties.Name],
                seeds: [
                    {
                        type: SeedTypeObjectId,
                        value: objectId,
                    },
                ],
            });
        });
    });

    it('renders legacy decoy membership when tier management is disabled', async () => {
        featureFlagEnabled = false;
        legacyMembers = [{ object_id: objectId }];

        render(
            <PotentialDecoyBanner
                nodeType={ActiveDirectoryNodeKind.User}
                objectId={objectId}
                properties={oldPotentialDecoyProperties}
            />
        );

        expect(await screen.findByRole('status')).toHaveTextContent(/marked as a decoy/i);
        await waitFor(() => {
            expect(legacyMembersQueryParams?.get('object_id')).toBe(`eq:${objectId}`);
        });
    });

    it('deletes the AGT selector when unmarking a decoy', async () => {
        agtSelectors = [
            {
                id: 12,
                asset_group_tag_id: decoyTag.id,
                name: oldPotentialDecoyProperties[CommonKindProperties.Name],
                seeds: [{ type: SeedTypeObjectId, value: objectId }],
            } as AssetGroupTagSelector,
        ];

        render(
            <PotentialDecoyBanner
                nodeType={ActiveDirectoryNodeKind.User}
                objectId={objectId}
                properties={oldPotentialDecoyProperties}
            />
        );

        const markDecoySwitch = await screen.findByRole('checkbox', {
            name: /mark object as decoy/i,
        });

        await waitFor(() => {
            expect(markDecoySwitch).toBeChecked();
            expect(markDecoySwitch).not.toBeDisabled();
        });

        await userEvent.click(markDecoySwitch);

        await waitFor(() => {
            expect(deletedSelectorId).toBe('12');
        });
    });

    it('removes the legacy selector when unmarking a decoy', async () => {
        featureFlagEnabled = false;
        legacyMembers = [{ object_id: objectId }];

        render(
            <PotentialDecoyBanner
                nodeType={ActiveDirectoryNodeKind.User}
                objectId={objectId}
                properties={oldPotentialDecoyProperties}
            />
        );

        const markDecoySwitch = await screen.findByRole('checkbox', {
            name: /mark object as decoy/i,
        });

        await waitFor(() => {
            expect(markDecoySwitch).toBeChecked();
            expect(markDecoySwitch).not.toBeDisabled();
        });

        await userEvent.click(markDecoySwitch);

        await waitFor(() => {
            expect(legacySelectorChangeset).toEqual([
                {
                    selector_name: objectId,
                    sid: objectId,
                    action: 'remove',
                },
            ]);
        });
    });

    it('hides the toggle without graph write permission', async () => {
        graphWritePermission = false;

        render(
            <PotentialDecoyBanner
                kinds={[TAG_DECOY_AGT]}
                nodeType={ActiveDirectoryNodeKind.User}
                objectId={objectId}
                properties={oldPotentialDecoyProperties}
            />
        );

        expect(await screen.findByRole('status')).toHaveTextContent(/marked as a decoy/i);
        expect(screen.queryByRole('checkbox', { name: /mark object as decoy/i })).not.toBeInTheDocument();
    });
});
