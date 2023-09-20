import { Box, Grid, useTheme } from '@mui/material';
import MenuContainer from '../Explore/Search/Menu/MenuContainer';
import EntityInfoPanel from '../Explore/EntityInfo/EntityInfoPanel';
import { AssetGroupMemberList, apiClient } from 'bh-shared-ui';
import { useQuery } from 'react-query';
import { SelectedNode } from 'src/ducks/entityinfo/types';
import { useState } from 'react';

const SetManagement = () => {
    const theme = useTheme();
    const [selectedNode, setSelectedNode] = useState<SelectedNode | null>(null);

    const listAssetGroupMembersQuery = useQuery(
        ["listAssetGroupMembers"],
        () => apiClient.listAssetGroupMembers("1").then(res => res.data)
    );

    const handleSelectMember = (member: any) => {
        setSelectedNode({
            id: member.object_id,
            type: member.primary_kind,
            name: member.name
        });
    }
    
    return (
        <Box height={"100%"} padding={theme.spacing(2, 4)}>
            <Grid container height={"100%"} spacing={2}>
                <Grid item xs={4} xl={3}>
                    <MenuContainer />
                </Grid>
                <Grid height={"100%"} item xs={4} xl={6}>
                    <AssetGroupMemberList 
                        assetGroupMembers={listAssetGroupMembersQuery.data?.data.members || []}
                        onSelectMember={handleSelectMember}
                    />
                </Grid>
                <Grid item xs={4} xl={3}>
                    <EntityInfoPanel selectedNode={selectedNode} />
                </Grid>
            </Grid>
        </Box>
    );
};

export default SetManagement;
