import { Box, Grid, Paper, useTheme } from '@mui/material';
import MenuContainer from '../Explore/Search/Menu/MenuContainer';
import EntityInfoPanel from '../Explore/EntityInfo/EntityInfoPanel';
import { AssetGroupMemberList, apiClient } from 'bh-shared-ui';
import { useQuery } from 'react-query';
import { setEntityInfoOpen, setSelectedNode } from 'src/ducks/entityinfo/actions';
import { useAppDispatch } from 'src/store';

const SetManagement = () => {

    const dispatch = useAppDispatch();
    const theme = useTheme();

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
                    <EntityInfoPanel />
                </Grid>
            </Grid>
        </Box>
    );
};

export default SetManagement;
