import { Box, Grid, Paper } from '@mui/material';
import MenuContainer from '../Explore/Search/Menu/MenuContainer';
import EntityInfoPanel from '../Explore/EntityInfo/EntityInfoPanel';
import { AssetGroupMemberList, apiClient } from 'bh-shared-ui';
import { useQuery } from 'react-query';
import { setEntityInfoOpen, setSelectedNode } from 'src/ducks/entityinfo/actions';
import { useAppDispatch } from 'src/store';

const SetManagement = () => {

    const dispatch = useAppDispatch();

    const listAssetGroupMembersQuery = useQuery(
        ["listAssetGroupMembers"],
        () => apiClient.listAssetGroupMembers("1").then(res => res.data)
    );

    const handleSelectMember = (member: any) => {
        dispatch(setEntityInfoOpen(true));
        dispatch(setSelectedNode({
            id: member.object_id,
            type: member.primary_kind,
            name: member.name
        }))
    }
    
    return (
        <Grid container height={"100%"} sx={{ padding: "32px 0 32px 16px" }}>
            <Grid item xs={3} sx={{ padding: "0 16px" }}>
                <Box component={Paper} height={350}>
                    <MenuContainer />
                </Box>
            </Grid>
            <Grid item xs={6} sx={{ padding: "0 16px" }}>
                <Box component={Paper} height={350}>
                    <AssetGroupMemberList 
                        assetGroupMembers={listAssetGroupMembersQuery.data?.data.members || []}
                        onSelectMember={handleSelectMember}
                    />
                </Box>
            </Grid>
            <Grid item xs={3}>
                <EntityInfoPanel />
            </Grid>
        </Grid>
    );
};

export default SetManagement;
