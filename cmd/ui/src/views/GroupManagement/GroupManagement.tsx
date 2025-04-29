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

import { faGem } from '@fortawesome/free-solid-svg-icons';
import {
    DropdownOption,
    EntityKinds,
    ExploreQueryParams,
    GroupManagementContent,
    HIGH_VALUE_LABEL,
    Permission,
    TIER_ZERO_TAG,
    createTypedSearchParams,
    searchbarActions,
    useAppNavigate,
    useExploreParams,
    useFeatureFlag,
    useInitialEnvironment,
    useNodeByObjectId,
    usePermissions,
} from 'bh-shared-ui';
import { AssetGroup, AssetGroupMember } from 'js-client-library';
import { useState } from 'react';
import { setSelectedNode } from 'src/ducks/entityinfo/actions';
import { SelectedNode } from 'src/ducks/entityinfo/types';
import { ROUTE_EXPLORE } from 'src/routes/constants';
import { useAppDispatch } from 'src/store';
import EntityInfoPanel from '../Explore/EntityInfo/EntityInfoPanel';
import { dataCollectionMessage } from '../QA/utils';

const GroupManagement = () => {
    const dispatch = useAppDispatch();
    const navigate = useAppNavigate();
    const backButtonFlagQuery = useFeatureFlag('back_button_support');

    const { data: environment } = useInitialEnvironment({ orderBy: 'name' });

    // Kept out of the shared UI due to diff between GraphNodeTypes across apps
    const [openNode, setOpenNode] = useState<SelectedNode | null>(null);
    const getGraphNodeByObjectId = useNodeByObjectId(openNode?.id);
    const { setExploreParams } = useExploreParams();

    const { checkPermission } = usePermissions();

    const handleClickMember = (member: AssetGroupMember) => {
        setOpenNode({
            id: member.object_id,
            type: member.primary_kind as EntityKinds,
            name: member.name,
        });
        if (backButtonFlagQuery.data?.enabled) {
            setExploreParams({ expandedPanelSections: null });
        }
    };

    const handleShowNodeInExplore = () => {
        if (openNode) {
            const searchNode = {
                objectid: openNode.id,
                label: openNode.name,
                ...openNode,
            };

            if (backButtonFlagQuery.data?.enabled) {
                navigate({
                    pathname: ROUTE_EXPLORE,
                    search: createTypedSearchParams<ExploreQueryParams>({
                        selectedItem: getGraphNodeByObjectId.data?.id,
                        primarySearch: openNode?.id,
                        searchType: 'node',
                        exploreSearchTab: 'node',
                    }),
                });
            } else {
                dispatch(searchbarActions.sourceNodeSelected(searchNode));
                dispatch(setSelectedNode(openNode));
                navigate(ROUTE_EXPLORE);
            }
        }
    };

    // Handle tier zero case
    const mapAssetGroups = (assetGroups: AssetGroup[]): DropdownOption[] => {
        return assetGroups.map((assetGroup) => {
            const isTierZero = assetGroup.tag === TIER_ZERO_TAG;
            return {
                key: assetGroup.id,
                value: isTierZero ? HIGH_VALUE_LABEL : assetGroup.name,
                icon: isTierZero ? faGem : undefined,
            };
        });
    };

    return (
        <GroupManagementContent
            globalEnvironment={environment ?? null}
            showExplorePageLink={!!openNode}
            tierZeroLabel={HIGH_VALUE_LABEL}
            tierZeroTag={TIER_ZERO_TAG}
            // Both these components should eventually be moved into the shared UI library
            entityPanelComponent={<EntityInfoPanel selectedNode={openNode} />}
            domainSelectorErrorMessage={<>Domains unavailable. {dataCollectionMessage}</>}
            onShowNodeInExplore={handleShowNodeInExplore}
            onClickMember={handleClickMember}
            mapAssetGroups={mapAssetGroups}
            userHasEditPermissions={checkPermission(Permission.GRAPH_DB_WRITE)}
        />
    );
};

export default GroupManagement;
