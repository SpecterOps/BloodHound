import { Box, Grid, Paper, Typography, useTheme } from '@mui/material';
import EntityInfoPanel from '../Explore/EntityInfo/EntityInfoPanel';
import { AssetGroupMemberList, apiClient, DropdownSelector, DropdownOption, AssetGroupEdit } from 'bh-shared-ui';
import { useQuery } from 'react-query';
import { SelectedNode } from 'src/ducks/entityinfo/types';
import { useEffect, useState } from 'react';
import DataSelector from '../QA/DataSelector';
import { AssetGroup, AssetGroupMember } from 'js-client-library';
import { GraphNodeTypes } from 'src/ducks/graph/types';
import { faGem } from '@fortawesome/free-solid-svg-icons';
import { useSelector } from 'react-redux';
import { Domain } from 'src/ducks/global/types';

type SelectedDomain = {
    id: string | null;
    type: string | null;
}

const SetManagement = () => {
    const theme = useTheme();

    const domain: Domain = useSelector((state: any) => state.global.options.domain);

    const [selectedDomain, setSelectedDomain] = useState<SelectedDomain | null>(null);
    const [selectedAssetGroup, setSelectedAssetGroup] = useState<AssetGroup | null>(null);
    const [assetGroupMembers, setAssetGroupMembers] = useState<AssetGroupMember[]>([]);
    const [selectedNode, setSelectedNode] = useState<SelectedNode | null>(null);
    
    
    const listAssetGroups = useQuery(
        ["listAssetGroups"],
        () => apiClient.listAssetGroups().then(res => res.data.data.asset_groups),
    );

    const listAssetGroupMembersQuery = useQuery(
        ["listAssetGroupMembers"],
        () => apiClient.listAssetGroupMembers(`${selectedAssetGroup?.id}`).then(res => res.data.data.members),
        { enabled: !!selectedAssetGroup }
    );

    useEffect(() => {
        const filterDomain = selectedDomain || domain;
        const filteredAssetGroupMembers = listAssetGroupMembersQuery.data?.filter(member => {
            switch (filterDomain.type) {
                case 'active-directory-platform':
                    return member.environment_kind === "Domain";
                case 'azure-platform':
                    return member.environment_kind === "Tenant";
                default:
                    return member.environment_id === filterDomain.id;
            }
        });
        setAssetGroupMembers(filteredAssetGroupMembers || []);
    }, [listAssetGroupMembersQuery.data, selectedDomain, domain])

    const handleSelectMember = (member: AssetGroupMember) => {
        setSelectedNode({
            id: member.object_id,
            type: member.primary_kind as GraphNodeTypes,
            name: member.name
        });
    }

    const handleAssetGroupSelectorChange = (selectedAssetGroup: DropdownOption) => {
        const selected = listAssetGroups.data?.find(assetGroup => assetGroup.id === selectedAssetGroup.key);
        if (selected) setSelectedAssetGroup(selected);
    }
    
    return (
        <Box height={"100%"} padding={theme.spacing(2, 4)}>
            <Grid container height={"100%"} spacing={2}>
                <Grid item xs={3} md={3}>
                    <Box component={Paper} elevation={0} marginBottom={1}>
                        <Grid container>
                            <Grid item xs={3} display={"flex"} alignItems={"center"} paddingLeft={1}>
                                <Typography variant="button">Group:</Typography>
                            </Grid>
                            <Grid item xs={9}>
                                <DropdownSelector
                                    options={listAssetGroups.data?.map((assetGroup: AssetGroup) => {
                                        return { key: assetGroup.id, value: assetGroup.name, icon: faGem };
                                    }) || []}
                                    selectedText={selectedAssetGroup?.name || "Loading..."}
                                    fullWidth
                                    onChange={handleAssetGroupSelectorChange}
                                />
                            </Grid>
                            <Grid item xs={3} display={"flex"} alignItems={"center"} paddingLeft={1}>
                                <Typography variant="button">Tenant:</Typography>
                            </Grid>
                            <Grid item xs={9}>
                                <DataSelector
                                    value={selectedDomain || domain || { type: null, id: null }}
                                    onChange={selection => setSelectedDomain({ ...selection })}
                                    fullWidth
                                />
                            </Grid>
                        </Grid>
                    </Box>
                    <AssetGroupEdit assetGroupId={selectedAssetGroup?.id.toString()} members={assetGroupMembers} />
                </Grid>
                <Grid height={"100%"} item xs={5} md={6}>
                    <AssetGroupMemberList 
                        assetGroupMembers={assetGroupMembers}
                        onSelectMember={handleSelectMember}
                    />
                </Grid>
                <Grid item xs={4} md={3}>
                    <EntityInfoPanel selectedNode={selectedNode} />
                </Grid>
            </Grid>
        </Box>
    );
};

export default SetManagement;
