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

import { AssetGroup, AssetGroupMember, AssetGroupMemberParams } from 'js-client-library';
import { FC, ReactNode, useEffect, useState } from 'react';
import DropdownSelector, { DropdownOption } from '../DropdownSelector';
import { Box, Button, Grid, Paper, Typography, useTheme } from '@mui/material';
import { useQuery } from 'react-query';
import { apiClient } from '../../utils';
import { faExternalLink } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import AssetGroupEdit from '../AssetGroupEdit';
import AssetGroupMemberList from '../AssetGroupMemberList';
import { SelectedDomain } from './types';
import DataSelector from '../../views/DataQuality/DataSelector';
import AssetGroupFilters from '../AssetGroupFilters';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../..';
import { FILTERABLE_PARAMS } from '../AssetGroupFilters/AssetGroupFilters';

// Top level layout and shared logic for the Group Management page
const GroupManagementContent: FC<{
    globalDomain: SelectedDomain;
    showExplorePageLink: boolean;
    tierZeroLabel: string;
    tierZeroTag: string;
    entityPanelComponent: ReactNode;
    domainSelectorErrorMessage: ReactNode;
    onShowNodeInExplore: () => void;
    onClickMember: (member: AssetGroupMember) => void;
    mapAssetGroups: (assetGroups: AssetGroup[]) => DropdownOption[];
}> = ({
    globalDomain,
    showExplorePageLink,
    tierZeroLabel,
    tierZeroTag,
    entityPanelComponent,
    domainSelectorErrorMessage,
    onShowNodeInExplore,
    onClickMember,
    mapAssetGroups,
}) => {
    const theme = useTheme();

    const [selectedDomain, setSelectedDomain] = useState<SelectedDomain | null>(null);
    const [selectedAssetGroupId, setSelectedAssetGroupId] = useState<number | null>(null);
    const [filterParams, setFilterParams] = useState<AssetGroupMemberParams>({});
    const [availableNodeKinds, setAvailableNodeKinds] = useState<Array<AzureNodeKind | ActiveDirectoryNodeKind>>([]);

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

    const makeNodeFilterable = (node: ActiveDirectoryNodeKind | AzureNodeKind) => {
        if (availableNodeKinds.includes(node)) return;
        setAvailableNodeKinds((prev) => [...prev, node]);
    };

    // Start building a filter query for members that gets passed down to AssetGroupMemberList to make the request
    useEffect(() => {
        const filterDomain = selectedDomain || globalDomain;
        const filter: AssetGroupMemberParams = {};
        if (filterDomain?.type === 'active-directory-platform') {
            filter.environment_kind = 'eq:Domain';
        } else if (filterDomain?.type === 'azure-platform') {
            filter.environment_kind = 'eq:AZTenant';
        } else {
            filter.environment_id = `eq:${filterDomain?.id}`;
        }
        setAvailableNodeKinds([]);
        setFilterParams(filter);
    }, [selectedDomain, globalDomain, selectedAssetGroupId]);

    const selectorLabelStyles = { display: { xs: 'none', xl: 'flex' } };

    return (
        <Box height={'100%'} padding={theme.spacing(2, 4)}>
            <Grid container height={'100%'} spacing={2}>
                <Grid item xs={3} md={3}>
                    <Box component={Paper} elevation={0} marginBottom={1}>
                        <Grid container>
                            <Grid item sm={4} sx={selectorLabelStyles} alignItems={'center'} paddingLeft={3}>
                                <Typography variant='button'>Group:</Typography>
                            </Grid>
                            <Grid item xs={12} xl={8}>
                                <DropdownSelector
                                    options={listAssetGroups.data ? mapAssetGroups(listAssetGroups.data) : []}
                                    selectedText={getAssetGroupSelectorLabel()}
                                    onChange={handleAssetGroupSelectorChange}
                                    fullWidth
                                />
                            </Grid>
                            <Grid item xs={4} sx={selectorLabelStyles} alignItems={'center'} paddingLeft={3}>
                                <Typography variant='button'>Environment:</Typography>
                            </Grid>
                            <Grid item xs={12} xl={8}>
                                <DataSelector
                                    value={selectedDomain || globalDomain || { type: null, id: null }}
                                    errorMessage={domainSelectorErrorMessage}
                                    onChange={(selection: SelectedDomain) => setSelectedDomain({ ...selection })}
                                    fullWidth={true}
                                />
                            </Grid>
                        </Grid>
                    </Box>
                    <AssetGroupFilters
                        filterParams={filterParams}
                        handleFilterChange={handleFilterChange}
                        availableNodeKinds={availableNodeKinds}
                    />
                    {selectedAssetGroup && (
                        <AssetGroupEdit
                            assetGroup={selectedAssetGroup}
                            filter={filterParams}
                            makeNodeFilterable={makeNodeFilterable}
                        />
                    )}
                </Grid>
                <Grid height={'100%'} item xs={5} md={6}>
                    <AssetGroupMemberList
                        assetGroup={selectedAssetGroup}
                        filter={filterParams}
                        onSelectMember={onClickMember}
                        canFilterToEmpty={!!availableNodeKinds.length}
                    />
                </Grid>
                <Grid item xs={4} md={3} height={'100%'}>
                    {/* CSS calc accounts for the height of the link button */}
                    <Box sx={{ maxHeight: 'calc(100% - 45px)', overflow: 'auto' }}>{entityPanelComponent}</Box>
                    {showExplorePageLink && (
                        <Button
                            data-testid='group-management_explore-link'
                            variant='contained'
                            disableElevation
                            fullWidth
                            sx={{ borderRadius: '4px', marginTop: '8px' }}
                            onClick={onShowNodeInExplore}
                            startIcon={<FontAwesomeIcon icon={faExternalLink} />}>
                            Open in Explore
                        </Button>
                    )}
                </Grid>
            </Grid>
        </Box>
    );
};

export default GroupManagementContent;
