import EntityInfoPanel from '../Explore/EntityInfo/EntityInfoPanel';
import { DropdownOption, GroupManagementContent } from 'bh-shared-ui';
import { SelectedNode } from 'src/ducks/entityinfo/types';
import { useState } from 'react';
import DataSelector from '../QA/DataSelector';
import { AssetGroup, AssetGroupMember } from 'js-client-library';
import { GraphNodeTypes } from 'src/ducks/graph/types';
import { faGem } from '@fortawesome/free-solid-svg-icons';
import { useDispatch, useSelector } from 'react-redux';
import { Domain } from 'src/ducks/global/types';
import { setSelectedNode } from 'src/ducks/entityinfo/actions';
import { useNavigate } from 'react-router-dom';
import { ROUTE_EXPLORE } from 'src/ducks/global/routes';
import { setSearchValue, startSearchSelected } from 'src/ducks/searchbar/actions';
import { PRIMARY_SEARCH, SEARCH_TYPE_EXACT } from 'src/ducks/searchbar/types';
import { TIER_ZERO_LABEL, TIER_ZERO_TAG } from 'src/constants';

const GroupManagement = () => {
    const dispatch = useDispatch();
    const navigate = useNavigate();

    const globalDomain: Domain = useSelector((state: any) => state.global.options.domain);

    // Kept out of the shared UI due to diff between GraphNodeTypes across apps
    const [openNode, setOpenNode] = useState<SelectedNode | null>(null);

    const handleClickMember = (member: AssetGroupMember) => {
        setOpenNode({
            id: member.object_id,
            type: member.primary_kind as GraphNodeTypes,
            name: member.name,
        });
    };

    const handleShowNodeInExplore = () => {
        if (openNode) {
            const searchNode = {
                objectid: openNode.id,
                label: openNode.name,
                ...openNode,
            };
            dispatch(setSearchValue(searchNode, PRIMARY_SEARCH, SEARCH_TYPE_EXACT));
            dispatch(startSearchSelected(PRIMARY_SEARCH));
            dispatch(setSelectedNode(openNode));

            navigate(ROUTE_EXPLORE);
        }
    };

    // Handle tier zero case
    const mapAssetGroups = (assetGroups: AssetGroup[]): DropdownOption[] => {
        return assetGroups.map((assetGroup) => {
            const isTierZero = assetGroup.tag === TIER_ZERO_TAG;
            return {
                key: assetGroup.id,
                value: isTierZero ? TIER_ZERO_LABEL : assetGroup.name,
                icon: isTierZero ? faGem : undefined,
            };
        });
    };

    return (
        <GroupManagementContent
            globalDomain={globalDomain}
            showExplorePageLink={!!openNode}
            tierZeroLabel={TIER_ZERO_LABEL}
            tierZeroTag={TIER_ZERO_TAG}
            // Both these components should eventually be moved into the shared UI library
            entityPanelComponent={<EntityInfoPanel selectedNode={openNode} />}
            generateDomainSelectorComponent={(props) => <DataSelector {...props} />}
            onShowNodeInExplore={handleShowNodeInExplore}
            onClickMember={handleClickMember}
            mapAssetGroups={mapAssetGroups}
        />
    );
};

export default GroupManagement;
