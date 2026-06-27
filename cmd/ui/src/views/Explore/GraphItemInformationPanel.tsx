// Copyright 2025 Specter Ops, Inc.
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

import {
    EdgeInfoPane,
    EntityInfoDataTableGraphed,
    EntityInfoPanel,
    EntityKinds,
    EntityTables,
    isEdge,
    isNode,
    useExploreSelectedItem,
} from 'bh-shared-ui';
import { HTMLProps } from 'react';
import { RACFGroupMembers, RACFGroupSubgroups, RACFUserGroups } from 'src/racfhound/RACFGroupMembers';
import {
    isRACFGroupKind,
    isRACFUserKind,
    RACF_GROUP_MEMBERS_SECTION,
    RACF_GROUP_SUBGROUPS_SECTION,
    RACF_USER_GROUPS_SECTION,
} from 'src/racfhound/groupMembers';

const defaultClasses: HTMLProps<HTMLElement>['className'] = 'bottom-0 top-0 py-4 absolute right-4';

const getRACFTables = (nodeType: string, databaseId: string): EntityTables | undefined => {
    if (isRACFGroupKind(nodeType)) {
        return [
            {
                sectionProps: {
                    id: databaseId,
                    label: RACF_GROUP_MEMBERS_SECTION,
                },
                TableComponent: RACFGroupMembers,
            },
            {
                sectionProps: {
                    id: databaseId,
                    label: RACF_GROUP_SUBGROUPS_SECTION,
                },
                TableComponent: RACFGroupSubgroups,
            },
        ];
    }

    if (isRACFUserKind(nodeType)) {
        return [
            {
                sectionProps: {
                    id: databaseId,
                    label: RACF_USER_GROUPS_SECTION,
                },
                TableComponent: RACFUserGroups,
            },
        ];
    }

    return undefined;
};

const GraphItemInformationPanel = () => {
    const { selectedItem, selectedItemQuery } = useExploreSelectedItem();

    if (!selectedItem || selectedItemQuery.isLoading) {
        return null;
    }

    if (selectedItemQuery.isError) {
        return (
            <EntityInfoPanel
                DataTable={EntityInfoDataTableGraphed}
                className={defaultClasses}
                selectedNode={{ graphId: selectedItem, id: '', name: 'Unknown', type: 'Unknown' as EntityKinds }}
            />
        );
    }

    if (selectedItemQuery.data && isEdge(selectedItemQuery.data)) {
        const selectedEdge = {
            id: selectedItem as string,
            name: selectedItemQuery.data.label || '',
            data: selectedItemQuery.data.properties || {},
            sourceNode: {
                id: selectedItemQuery.data.source,
                objectId: selectedItemQuery.data.sourceNode.objectId,
                name: selectedItemQuery.data.sourceNode.label,
                type: selectedItemQuery.data.sourceNode.kind,
            },
            targetNode: {
                id: selectedItemQuery.data.target,
                objectId: selectedItemQuery.data.targetNode.objectId,
                name: selectedItemQuery.data.targetNode.label,
                type: selectedItemQuery.data.targetNode.kind,
            },
        };
        return <EdgeInfoPane className={defaultClasses} selectedEdge={selectedEdge} />;
    }

    if (selectedItemQuery.data && isNode(selectedItemQuery.data)) {
        const selectedNode = {
            graphId: selectedItem,
            id: selectedItemQuery.data.objectId,
            name: selectedItemQuery.data.label,
            type: selectedItemQuery.data.kind as EntityKinds,
        };
        const additionalTables = getRACFTables(selectedItemQuery.data.kind, selectedItem);

        return (
            <EntityInfoPanel
                className={defaultClasses}
                selectedNode={selectedNode}
                DataTable={EntityInfoDataTableGraphed}
                additionalTables={additionalTables}
            />
        );
    }
};

export default GraphItemInformationPanel;
