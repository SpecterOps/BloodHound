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

type SelectedDomain = {
    id: string | null;
    type: string | null;
};

const GroupManagementContent: FC<{
    globalDomain: SelectedDomain;
    showExplorePageLink: boolean;
    entityPanelComponent: ReactNode;
    generateDomainSelectorComponent: (props: any) => ReactNode;
    onShowNodeInExplore: () => void;
    onClickMember: (member: AssetGroupMember) => void;
    mapAssetGroups: (assetGroups: AssetGroup[]) => DropdownOption[];
}> = ({
    globalDomain,
    showExplorePageLink,
    entityPanelComponent,
    generateDomainSelectorComponent,
    onShowNodeInExplore,
    onClickMember,
    mapAssetGroups,
}) => {
    const theme = useTheme();

    const [selectedDomain, setSelectedDomain] = useState<SelectedDomain | null>(null);
    const [selectedAssetGroup, setSelectedAssetGroup] = useState<AssetGroup | null>(null);
    const [filterParams, setFilterParams] = useState<AssetGroupMemberParams>({});

    const setInitialGroup = (data: AssetGroup[]) => {
        if (!selectedAssetGroup && data.length) {
            const initialGroup = data.find((group) => group.tag === 'admin_tier_zero') || data[0];
            setSelectedAssetGroup(initialGroup);
        }
    };

    const listAssetGroups = useQuery(
        ['listAssetGroups'],
        () => apiClient.listAssetGroups().then((res) => res.data.data.asset_groups),
        { onSuccess: setInitialGroup }
    );

    const handleAssetGroupSelectorChange = (selectedAssetGroup: DropdownOption) => {
        const selected = listAssetGroups.data?.find((assetGroup) => assetGroup.id === selectedAssetGroup.key);
        if (selected) setSelectedAssetGroup(selected);
    };

    const getDomainSelectorProps = () => {
        return {
            value: selectedDomain || globalDomain || { type: null, id: null },
            onChange: (selection: SelectedDomain) => setSelectedDomain({ ...selection }),
            fullWidth: true,
        };
    };

    useEffect(() => {
        const filterDomain = selectedDomain || globalDomain;
        const filter: AssetGroupMemberParams = {};
        if (filterDomain.type === 'active-directory-platform') {
            filter.environment_kind = 'eq:Domain';
        } else if (filterDomain.type === 'azure-platform') {
            filter.environment_kind = 'eq:AZTenant';
        } else {
            filter.environment_id = `eq:${filterDomain.id}`;
        }
        setFilterParams(filter);
    }, [selectedDomain, globalDomain, selectedAssetGroup]);

    return (
        <Box height={'100%'} padding={theme.spacing(2, 4)}>
            <Grid container height={'100%'} spacing={2}>
                <Grid item xs={3} md={3}>
                    <Box component={Paper} elevation={0} marginBottom={1}>
                        <Grid container>
                            <Grid item xs={3} display={'flex'} alignItems={'center'} paddingLeft={1}>
                                <Typography variant='button'>Group:</Typography>
                            </Grid>
                            <Grid item xs={9}>
                                <DropdownSelector
                                    options={listAssetGroups.data ? mapAssetGroups(listAssetGroups.data) : []}
                                    selectedText={selectedAssetGroup?.name || 'Loading...'}
                                    onChange={handleAssetGroupSelectorChange}
                                    fullWidth
                                />
                            </Grid>
                            <Grid item xs={3} display={'flex'} alignItems={'center'} paddingLeft={1}>
                                <Typography variant='button'>Tenant:</Typography>
                            </Grid>
                            <Grid item xs={9}>
                                {generateDomainSelectorComponent(getDomainSelectorProps())}
                            </Grid>
                        </Grid>
                    </Box>
                    {selectedAssetGroup && <AssetGroupEdit assetGroup={selectedAssetGroup} filter={filterParams} />}
                </Grid>
                <Grid height={'100%'} item xs={5} md={6}>
                    <AssetGroupMemberList
                        assetGroup={selectedAssetGroup}
                        filter={filterParams}
                        onSelectMember={onClickMember}
                    />
                </Grid>
                <Grid item xs={4} md={3} height={'100%'}>
                    <Box sx={{ maxHeight: 'calc(100% - 45px)', overflow: 'auto' }}>{entityPanelComponent}</Box>
                    {showExplorePageLink && (
                        <Button
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
