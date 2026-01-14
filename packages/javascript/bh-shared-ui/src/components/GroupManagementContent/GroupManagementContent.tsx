// Copyright 2023 Specter Ops, Inc.
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

import { Badge, Button } from '@bloodhoundenterprise/doodleui';
import { faExternalLink, faEyeSlash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Grid, Typography } from '@mui/material';
import { AssetGroup, AssetGroupMember, AssetGroupMemberParams } from 'js-client-library';
import { FC, HTMLProps, ReactNode, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import useRoleBasedFiltering from '../../hooks/useRoleBasedFiltering';
import { apiClient } from '../../utils/api';
import AssetGroupEdit from '../AssetGroupEdit/AssetGroupEdit';
import AssetGroupFilters from '../AssetGroupFilters';
import { FILTERABLE_PARAMS } from '../AssetGroupFilters/AssetGroupFilters';
import AssetGroupMemberList from '../AssetGroupMemberList';
import DropdownSelector, { DropdownOption } from '../DropdownSelector';
import { SelectedEnvironment, SimpleEnvironmentSelector } from '../SimpleEnvironmentSelector';

interface GroupManagementContentProps {
    globalEnvironment: SelectedEnvironment | null;
    showExplorePageLink: boolean;
    tierZeroLabel: string;
    tierZeroTag: string;
    entityPanelComponent: ReactNode;
    domainSelectorErrorMessage: ReactNode;
    onShowNodeInExplore: () => void;
    onClickMember: (member: AssetGroupMember) => void;
    mapAssetGroups: (assetGroups: AssetGroup[]) => DropdownOption[];
    userHasEditPermissions: boolean;
}

// Top level layout and shared logic for the Group Management page
const GroupManagementContent: FC<GroupManagementContentProps> = ({
    globalEnvironment,
    showExplorePageLink,
    tierZeroLabel,
    tierZeroTag,
    entityPanelComponent,
    domainSelectorErrorMessage,
    onShowNodeInExplore,
    onClickMember,
    mapAssetGroups,
    userHasEditPermissions,
}) => {
    const [selectedEnvironment, setSelectedEnvironment] = useState<SelectedEnvironment | null>(null);
    const [selectedAssetGroupId, setSelectedAssetGroupId] = useState<number | null>(null);
    const [filterParams, setFilterParams] = useState<AssetGroupMemberParams>({});

    const setInitialGroup = (data: AssetGroup[]) => {
        if (!selectedAssetGroupId && data?.length) {
            const initialGroup = data.find((group) => group.tag === tierZeroTag) || data[0];
            setSelectedAssetGroupId(initialGroup.id);
        }
    };

    const listAssetGroups = useQuery(
        ['listAssetGroups'],
        () => apiClient.listAssetGroups().then((res) => res.data.data.asset_groups),
        { onSuccess: setInitialGroup }
    );

    const selectedAssetGroup = listAssetGroups.data?.find((group) => group.id === selectedAssetGroupId) || null;

    const { data: memberCounts } = useQuery({
        queryKey: [
            'getAssetGroupMembersCount',
            filterParams.environment_id,
            filterParams.environment_kind,
            selectedAssetGroup,
        ],
        enabled: !!selectedAssetGroupId,
        queryFn: ({ signal }) =>
            apiClient
                .getAssetGroupMembersCount(
                    selectedAssetGroupId?.toString() ?? '', // This query will only execute if selectedAssetGroup is truthy.
                    { environment_id: filterParams.environment_id, environment_kind: filterParams.environment_kind },
                    { signal }
                )
                .then((res) => res.data.data),
    });

    const handleAssetGroupSelectorChange = (selectedAssetGroup: DropdownOption) => {
        const selected = listAssetGroups.data?.find((assetGroup) => assetGroup.id === selectedAssetGroup.key);
        if (selected) setSelectedAssetGroupId(selected.id);
    };

    const getAssetGroupSelectorLabel = (): string => {
        if (selectedAssetGroup?.tag === tierZeroTag) return tierZeroLabel;
        return selectedAssetGroup?.name || 'Select a Group';
    };

    const handleFilterChange = (key: (typeof FILTERABLE_PARAMS)[number], value: string) => {
        // Custom Member filter displays custom members, or all members.
        // If we want to also display only non customer members, change this:
        if (key === 'custom_member' && value.toLowerCase().includes('false')) {
            setFilterParams((prev) => {
                const _filterParams = { ...prev };
                delete _filterParams.custom_member;
                return _filterParams;
            });
            return;
        }
        setFilterParams((prev) => ({ ...prev, [key]: value.toString() }));
    };

    const handleSelect = (selection: SelectedEnvironment) => setSelectedEnvironment({ ...selection });

    // Start building a filter query for members that gets passed down to AssetGroupMemberList to make the request
    useEffect(() => {
        const filterDomain = selectedEnvironment || globalEnvironment;
        const filter: AssetGroupMemberParams = {};
        if (filterDomain?.type === 'active-directory-platform') {
            filter.environment_kind = 'eq:Domain';
        } else if (filterDomain?.type === 'azure-platform') {
            filter.environment_kind = 'eq:AZTenant';
        } else {
            filter.environment_id = `eq:${filterDomain?.id}`;
        }
        setFilterParams(filter);
    }, [selectedEnvironment, globalEnvironment, selectedAssetGroupId]);

    const selectorLabelStyles: HTMLProps<HTMLElement>['className'] = 'flex sm:hidden';

    const isRoleBasedFiltering = useRoleBasedFiltering();

    return (
        <div className='h-full py-4 px-8'>
            {isRoleBasedFiltering && (
                <Badge
                    data-testid='explore_entity-information-panel-badge-etac-filtering'
                    className='w-full justify-center text-sm text-neutral-dark-1 bg-[#F8EEFD] dark:bg-[#472E54] dark:text-neutral-light-1 border-0 mb-2'
                    icon={<FontAwesomeIcon icon={faEyeSlash} className='mr-2' />}
                    label='This account does not have access to this page. Please contact an administrator if this message is in error.'
                />
            )}
            <Grid container height={'100%'} spacing={2}>
                <Grid item xs={3} md={3}>
                    <div className='mb-2'>
                        <Grid container className='bg-neutral-2'>
                            <Grid item sm={4} className={selectorLabelStyles} alignItems={'center'} paddingLeft={3}>
                                <Typography variant='button'>Group:</Typography>
                            </Grid>
                            <Grid item xs={12} xl={8}>
                                <div className='p-2'>
                                    <DropdownSelector
                                        options={listAssetGroups.data ? mapAssetGroups(listAssetGroups.data) : []}
                                        selectedText={getAssetGroupSelectorLabel()}
                                        onChange={handleAssetGroupSelectorChange}
                                    />
                                </div>
                            </Grid>
                            <Grid item xs={4} className={selectorLabelStyles} alignItems={'center'} paddingLeft={3}>
                                <Typography variant='button'>Environment:</Typography>
                            </Grid>
                            <Grid item xs={12} xl={8} className='p-2'>
                                <SimpleEnvironmentSelector
                                    selected={selectedEnvironment || globalEnvironment || { type: null, id: null }}
                                    errorMessage={domainSelectorErrorMessage}
                                    buttonPrimary
                                    onSelect={handleSelect}
                                />
                            </Grid>
                        </Grid>
                    </div>
                    <AssetGroupFilters
                        filterParams={filterParams}
                        handleFilterChange={handleFilterChange}
                        memberCounts={memberCounts}
                    />
                    {selectedAssetGroup && (
                        <AssetGroupEdit
                            assetGroup={selectedAssetGroup}
                            filter={filterParams}
                            memberCounts={memberCounts}
                            isEditable={userHasEditPermissions}
                        />
                    )}
                </Grid>
                <Grid height={'100%'} item xs={5} md={6}>
                    <AssetGroupMemberList
                        assetGroup={selectedAssetGroup}
                        filter={filterParams}
                        onSelectMember={onClickMember}
                        canFilterToEmpty={(memberCounts?.total_count ?? 0) > 0}
                    />
                </Grid>
                <Grid item xs={4} md={3} height={'100%'}>
                    {/* CSS calc accounts for the height of the link button */}
                    <div className='max-h-[calc(100%-45px)] overflow-auto'>{entityPanelComponent}</div>
                    {showExplorePageLink && (
                        <Button
                            data-testid='group-management_explore-link'
                            style={{ borderRadius: '4px', marginTop: '8px', width: '100%' }}
                            onClick={onShowNodeInExplore}>
                            <FontAwesomeIcon icon={faExternalLink} />
                            <Typography ml='8px'>Open in Explore</Typography>
                        </Button>
                    )}
                </Grid>
            </Grid>
        </div>
    );
};

export default GroupManagementContent;
