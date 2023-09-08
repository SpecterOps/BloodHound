import { Box, Grid, Paper, useTheme } from '@mui/material';
import MenuContainer from '../Explore/Search/Menu/MenuContainer';
import EntityInfoPanel from '../Explore/EntityInfo/EntityInfoPanel';
import { useSelector } from 'react-redux';
import { AppState } from 'src/store';

const SetManagement = () => {

    const assetGroupState = useSelector((state: AppState) => state.assetgroups);

    console.log(assetGroupState);
    
    return (
        <Grid container sx={{ padding: "32px 0 32px 16px" }}>
            <Grid item xs={3} sx={{ padding: "0 16px" }}>
                <Box component={Paper} height={350}>
                    <MenuContainer />
                </Box>
            </Grid>
            <Grid item xs={6} sx={{ padding: "0 16px" }}>
                <Box component={Paper} height={350}>Set Data</Box>
            </Grid>
            <Grid item xs={3}>
                <EntityInfoPanel />
            </Grid>
        </Grid>
    );
};

export default SetManagement;
