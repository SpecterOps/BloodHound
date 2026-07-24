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
import { faWarning } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, CircularProgress, FormControlLabel, Switch, Typography } from '@mui/material';
import {
    AssetGroupTagTypeDecoy,
    SeedTypeObjectId,
    type AssetGroup,
    type AssetGroupTagSelector,
} from 'js-client-library';
import React from 'react';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { DECOY_OBJECT_TAG, TAG_DECOY_AGT } from '../../constants';
import { ActiveDirectoryKindProperties, ActiveDirectoryNodeKind, CommonKindProperties } from '../../graphSchema';
import { useFeatureFlag, usePermissions, useTagsQuery } from '../../hooks';
import { useNotifications } from '../../providers';
import { Permission, apiClient, cn } from '../../utils';

const POTENTIAL_DECOY_MIN_AGE_DAYS = 60;
const MS_PER_DAY = 24 * 60 * 60 * 1000;

const isUnsetLogonValue = (value: any): boolean => {
    return (
        value === undefined ||
        value === null ||
        value === '' ||
        value === 0 ||
        value === -1 ||
        value === '0' ||
        value === '-1'
    );
};

const isTruthyValue = (value: any): boolean => {
    return value === true || value === 1 || value === '1' || `${value}`.toLowerCase() === 'true';
};

const toEpochMilliseconds = (value: any): number | undefined => {
    if (typeof value === 'number' && Number.isFinite(value) && value > 0) {
        return value > 10_000_000_000 ? value : value * 1000;
    }

    if (typeof value !== 'string' || value.trim() === '') {
        return undefined;
    }

    const trimmedValue = value.trim();
    const numericValue = Number(trimmedValue);

    if (Number.isFinite(numericValue) && numericValue > 0) {
        return numericValue > 10_000_000_000 ? numericValue : numericValue * 1000;
    }

    const parsedDate = Date.parse(trimmedValue);

    return Number.isNaN(parsedDate) ? undefined : parsedDate;
};

const isAtLeastDaysOld = (value: any, days: number): boolean => {
    const timestamp = toEpochMilliseconds(value);

    if (timestamp === undefined) {
        return false;
    }

    return Date.now() - timestamp >= days * MS_PER_DAY;
};

const isExcludedSpecialAccount = (properties: Record<string, any>): boolean => {
    const accountIdentifiers = [
        properties[CommonKindProperties.ObjectID],
        properties[CommonKindProperties.Name],
        properties[ActiveDirectoryKindProperties.SamAccountName],
    ].filter((value): value is string => typeof value === 'string');

    return accountIdentifiers.some((value) => {
        const normalizedValue = value.toUpperCase();

        return (
            normalizedValue.endsWith('-500') ||
            normalizedValue.startsWith('KRBTGT') ||
            normalizedValue.startsWith('AZUREADKERBEROS.') ||
            normalizedValue.startsWith('AZUREADSSOACC.')
        );
    });
};

const isManagedServiceAccount = (properties: Record<string, any>): boolean => {
    return (
        isTruthyValue(properties[ActiveDirectoryKindProperties.GMSA]) ||
        isTruthyValue(properties[ActiveDirectoryKindProperties.MSA])
    );
};

export const isPotentialDecoyUser = (nodeType: string, properties: Record<string, any> = {}): boolean => {
    return (
        nodeType === ActiveDirectoryNodeKind.User &&
        isUnsetLogonValue(properties[ActiveDirectoryKindProperties.LastLogon]) &&
        isUnsetLogonValue(properties[ActiveDirectoryKindProperties.LastLogonTimestamp]) &&
        isTruthyValue(properties[CommonKindProperties.Enabled]) &&
        isAtLeastDaysOld(properties[CommonKindProperties.WhenCreated], POTENTIAL_DECOY_MIN_AGE_DAYS) &&
        !isExcludedSpecialAccount(properties) &&
        !isManagedServiceAccount(properties)
    );
};

const getDecoyAssetGroup = (assetGroups: AssetGroup[] | undefined): AssetGroup | undefined => {
    return assetGroups?.find((assetGroup) => assetGroup.tag === DECOY_OBJECT_TAG);
};

const hasLegacyDecoyTag = (properties: Record<string, any> = {}): boolean => {
    return properties[CommonKindProperties.SystemTags]?.includes(DECOY_OBJECT_TAG) === true;
};

const hasAgtDecoyTag = (kinds: string[] = []): boolean => {
    return kinds.includes(TAG_DECOY_AGT);
};

interface MatchingDecoySelectorResult {
    hasMatchingSelector: boolean;
    removableSelector?: AssetGroupTagSelector;
}

const isStandaloneObjectIdSelector = (selector: AssetGroupTagSelector, objectId: string): boolean => {
    const seed = selector.seeds[0];

    return (
        selector.is_default !== true &&
        selector.seeds.length === 1 &&
        seed.type === SeedTypeObjectId &&
        seed.value === objectId
    );
};

const PotentialDecoyBanner: React.FC<{
    kinds?: string[];
    nodeId?: number;
    nodeType: string;
    objectId: string;
    onDecoyUpdated?: () => void;
    properties?: Record<string, any>;
}> = ({ kinds = [], nodeId, nodeType, objectId, onDecoyUpdated = () => {}, properties = {} }) => {
    const [decoyStateOverride, setDecoyStateOverride] = React.useState<boolean>();
    const queryClient = useQueryClient();
    const { addNotification } = useNotifications();
    const { checkPermission } = usePermissions();
    const tierFlagQuery = useFeatureFlag('tier_management_engine');
    const tagsQuery = useTagsQuery();
    const hasPermission = checkPermission(Permission.GRAPH_DB_WRITE);
    const isTierManagementEnabled = tierFlagQuery.data?.enabled === true;
    const isLegacyAssetGroupEnabled = tierFlagQuery.isSuccess && tierFlagQuery.data?.enabled !== true;
    const potentialDecoy = isPotentialDecoyUser(nodeType, properties);
    const markedByGraph = hasAgtDecoyTag(kinds) || hasLegacyDecoyTag(properties);

    const decoyTag = tagsQuery.data?.find((tag) => tag.type === AssetGroupTagTypeDecoy);

    const legacyAssetGroupsQuery = useQuery(
        ['decoyAssetGroup'],
        () => apiClient.listAssetGroups().then((res) => getDecoyAssetGroup(res.data.data.asset_groups)),
        {
            enabled: isLegacyAssetGroupEnabled && hasPermission && (potentialDecoy || markedByGraph),
        }
    );

    const decoyAssetGroup = legacyAssetGroupsQuery.data;

    const legacyMembersQuery = useQuery(
        ['decoyAssetGroupMembers', decoyAssetGroup?.id, objectId],
        () =>
            apiClient
                .listAssetGroupMembers(decoyAssetGroup!.id, undefined, {
                    params: {
                        object_id: `eq:${objectId}`,
                    },
                })
                .then((res) => res.data.data.members),
        {
            enabled: isLegacyAssetGroupEnabled && hasPermission && !!decoyAssetGroup && !!objectId,
        }
    );

    const agtSelectorsQuery = useQuery(
        ['decoyAssetGroupTagSelector', decoyTag?.id, objectId],
        async (): Promise<MatchingDecoySelectorResult> => {
            const response = await apiClient.getAssetGroupTagSelectors(decoyTag!.id, {
                params: {
                    disabled_at: 'eq:null',
                    limit: 2,
                    type: `eq:${SeedTypeObjectId}`,
                    value: `eq:${objectId}`,
                },
            });
            const matchingSelectors = response.data.data.selectors ?? [];

            if (matchingSelectors.length !== 1) {
                return { hasMatchingSelector: matchingSelectors.length > 0 };
            }

            const selectorResponse = await apiClient.getAssetGroupTagSelector(decoyTag!.id, matchingSelectors[0].id);
            const selector = selectorResponse.data.data.selector;

            return {
                hasMatchingSelector: true,
                removableSelector: isStandaloneObjectIdSelector(selector, objectId) ? selector : undefined,
            };
        },
        {
            enabled:
                isTierManagementEnabled &&
                hasPermission &&
                !!decoyTag &&
                !!objectId &&
                (potentialDecoy || markedByGraph),
        }
    );

    const agtMembershipQuery = useQuery(
        ['decoyAssetGroupTagMembership', decoyTag?.id, nodeId],
        () =>
            apiClient
                .getAssetGroupTagMemberInfo(decoyTag!.id, nodeId!)
                .then((res) => res.data.data.member.selectors ?? []),
        {
            enabled: isTierManagementEnabled && hasPermission && !!decoyTag && nodeId !== undefined && markedByGraph,
        }
    );

    const removableSelector = agtSelectorsQuery.data?.removableSelector;
    const isOnlyMembershipSource =
        !markedByGraph ||
        (agtMembershipQuery.data?.length === 1 && agtMembershipQuery.data[0].id === removableSelector?.id);
    const canRemoveAgtDecoy = removableSelector !== undefined && isOnlyMembershipSource;
    const canRemoveLegacyDecoy =
        legacyMembersQuery.data?.length === 1 && legacyMembersQuery.data[0].custom_member === true;

    const serverMarkedDecoy =
        markedByGraph ||
        (legacyMembersQuery.data?.length ?? 0) > 0 ||
        agtSelectorsQuery.data?.hasMatchingSelector === true;
    const isMarkedDecoy = decoyStateOverride ?? serverMarkedDecoy;

    React.useEffect(() => {
        setDecoyStateOverride(undefined);
    }, [objectId]);

    React.useEffect(() => {
        if (decoyStateOverride !== undefined && decoyStateOverride === serverMarkedDecoy) {
            setDecoyStateOverride(undefined);
        }
    }, [decoyStateOverride, serverMarkedDecoy]);

    const toggleMutation = useMutation<unknown, unknown, boolean, { previousDecoyState: boolean | undefined }>(
        (checked: boolean) => {
            if (isTierManagementEnabled) {
                if (!decoyTag) return Promise.reject(new Error('Decoy tag not found'));

                if (checked) {
                    return apiClient.createAssetGroupTagSelector(decoyTag.id, {
                        name: properties[CommonKindProperties.Name] ?? objectId,
                        seeds: [
                            {
                                type: SeedTypeObjectId,
                                value: objectId,
                            },
                        ],
                    });
                }

                if (!canRemoveAgtDecoy || !removableSelector) {
                    return Promise.reject(new Error('Decoy membership must be removed from its rules'));
                }

                return apiClient.deleteAssetGroupTagSelector(decoyTag.id, removableSelector.id);
            }

            if (!decoyAssetGroup) return Promise.reject(new Error('Decoy asset group not found'));
            if (!checked && !canRemoveLegacyDecoy) {
                return Promise.reject(new Error('Decoy membership must be removed from its selectors'));
            }

            return apiClient.updateAssetGroupSelector(decoyAssetGroup.id, [
                {
                    selector_name: objectId,
                    sid: objectId,
                    action: checked ? 'add' : 'remove',
                },
            ]);
        },
        {
            onMutate: (checked) => {
                const previousDecoyState = decoyStateOverride;
                setDecoyStateOverride(checked);

                return { previousDecoyState };
            },
            onSuccess: (_data, checked) => {
                queryClient.invalidateQueries(['decoyAssetGroupMembers', decoyAssetGroup?.id, objectId]);
                queryClient.invalidateQueries(['decoyAssetGroupTagSelector', decoyTag?.id, objectId]);
                queryClient.invalidateQueries(['decoyAssetGroupTagMembership', decoyTag?.id, nodeId]);
                onDecoyUpdated();
                addNotification(
                    checked ? 'Object marked as decoy.' : 'Object unmarked as decoy.',
                    checked ? 'DecoyUpdateSuccess' : 'DecoyRemoveSuccess'
                );
            },
            onError: (error, _checked, context) => {
                setDecoyStateOverride(context?.previousDecoyState);
                console.error(error);
                addNotification('An error occurred when updating decoy status.', 'DecoyUpdateError');
            },
        }
    );

    if (!potentialDecoy && !isMarkedDecoy) return null;

    const isLoading =
        toggleMutation.isLoading ||
        tierFlagQuery.isLoading ||
        legacyAssetGroupsQuery.isLoading ||
        legacyMembersQuery.isLoading ||
        agtSelectorsQuery.isLoading ||
        agtMembershipQuery.isLoading;

    const canRemoveDecoy = isTierManagementEnabled ? canRemoveAgtDecoy : canRemoveLegacyDecoy;
    const requiresRuleManagement = isMarkedDecoy && !isLoading && !canRemoveDecoy;
    const toggleDisabled =
        isLoading ||
        !objectId ||
        (isTierManagementEnabled ? !decoyTag : !decoyAssetGroup) ||
        (isMarkedDecoy && !canRemoveDecoy);

    return (
        <Box
            role={isMarkedDecoy ? 'status' : 'alert'}
            className='mb-2 flex items-center justify-between gap-2 rounded border border-l-4 border-[#ed6c02] bg-[#fb923c]/15 px-2 py-1.5'>
            <Box alignItems='center' display='flex' gap={1} minWidth={0}>
                <Box
                    aria-hidden='true'
                    color='warning.dark'
                    component='span'
                    display='flex'
                    fontSize='0.875rem'
                    lineHeight={1}>
                    <FontAwesomeIcon icon={faWarning} />
                </Box>
                <Typography
                    color='text.primary'
                    variant='caption'
                    className={cn('leading-[1.3]', isMarkedDecoy ? 'font-semibold' : 'font-medium')}>
                    {isMarkedDecoy && requiresRuleManagement
                        ? 'This object is marked as a decoy by one or more rules. Manage its Decoy rules to remove it.'
                        : isMarkedDecoy
                          ? 'This object is marked as a decoy.'
                          : `No recorded AD logon, enabled, older than ${POTENTIAL_DECOY_MIN_AGE_DAYS} days; this user might be a decoy.`}
                </Typography>
            </Box>
            {hasPermission && (
                <Box alignItems='center' display='flex' flexShrink={0} gap={0.5}>
                    {isLoading && <CircularProgress size={14} />}
                    <FormControlLabel
                        control={
                            <Switch
                                checked={isMarkedDecoy}
                                disabled={toggleDisabled}
                                inputProps={{ 'aria-label': 'Mark object as decoy' }}
                                onChange={(_event, checked) => toggleMutation.mutate(checked)}
                                size='small'
                            />
                        }
                        label='Decoy'
                        classes={{
                            root: '!m-0',
                            label: '!text-xs !font-semibold',
                        }}
                    />
                </Box>
            )}
        </Box>
    );
};

export default PotentialDecoyBanner;
