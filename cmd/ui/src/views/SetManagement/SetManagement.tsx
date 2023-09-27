import { Box, Grid, Paper, Typography, useTheme } from '@mui/material';
import EntityInfoPanel from '../Explore/EntityInfo/EntityInfoPanel';
import { AssetGroupMemberList, apiClient, DropdownSelector, DropdownOption, AssetGroupAutocomplete } from 'bh-shared-ui';
import { useQuery } from 'react-query';
import { SelectedNode } from 'src/ducks/entityinfo/types';
import { useEffect, useState } from 'react';
import DataSelector from '../QA/DataSelector';
import { AssetGroup, AssetGroupMember } from 'js-client-library';
import { GraphNodeTypes } from 'src/ducks/graph/types';
import { faGem } from '@fortawesome/free-solid-svg-icons';
import { SubHeader } from '../Explore/fragments';
import { getKeywordAndTypeValues, useSearch } from 'src/hooks/useSearch';
import { useDebouncedValue } from 'src/hooks/useDebouncedValue';

type Domain = {
    type: string | null;
    id: string | null;
}

const SetManagement = () => {
    const theme = useTheme();
    const [selectedNode, setSelectedNode] = useState<SelectedNode | null>(null);
    const [domain, setDomain] = useState<Domain>({ type: null, id: null });
    const [selectedAssetGroup, setSelectedAssetGroup] = useState<AssetGroup | null>(null);
    const [assetGroupMembers, setAssetGroupMembers] = useState<AssetGroupMember[]>([]);

    const [searchInput, setSearchInput] = useState('');
    const debouncedInputValue = useDebouncedValue(searchInput, 250);
    const { keyword, type } = getKeywordAndTypeValues(debouncedInputValue);
    const search = useSearch(keyword, type);

    const handleInputChange = (_event: any, value: string) => {
        setSearchInput(value);
    }

    const listAssetGroups = useQuery(
        ["listAssetGroups"],
        () => apiClient.listAssetGroups().then(res => res.data.data.asset_groups),
    );

    const listAssetGroupMembersQuery = useQuery(
        ["listAssetGroupMembers"],
        () => apiClient.listAssetGroupMembers(selectedAssetGroup?.id.toString() || "1").then(res => res.data.data.members),
        { enabled: !!selectedAssetGroup?.id }
    );

    useEffect(() => {
        const filteredAssetGroupMembers = listAssetGroupMembersQuery.data?.filter(member => {
            switch (domain.type) {
                case 'active-directory-platform':
                    return member.environment_kind === "Domain";
                case 'azure-platform':
                    return member.environment_kind === "Tenant";
                default:
                    return member.environment_id === domain.id;
            }
        });
        setAssetGroupMembers(filteredAssetGroupMembers || []);
    }, [listAssetGroupMembersQuery.data, domain])

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
                            <Grid item xs={3} display={"flex"} alignItems={"center"} paddingLeft={3}>
                                <Typography variant="button">Set:</Typography>
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
                            <Grid item xs={3} display={"flex"} alignItems={"center"} paddingLeft={3}>
                                <Typography variant="button">Tenant:</Typography>
                            </Grid>
                            <Grid item xs={9}>
                                <DataSelector
                                    value={domain}
                                    onChange={setDomain}
                                    fullWidth
                                />
                            </Grid>
                        </Grid>
                    </Box>
                    <Box component={Paper} elevation={0} padding={1}>
                        <SubHeader label={'Total Members'} count={listAssetGroupMembersQuery.data?.length} />
                        <AssetGroupAutocomplete
                            search={search}
                            changelog={[]}
                            onInputChange={handleInputChange}
                            inputValue={searchInput}
                        />
                    </Box>
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
