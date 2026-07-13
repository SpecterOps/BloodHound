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

import { NodeDetails, RelationshipDetails } from 'js-client-library';
import { HTMLProps } from 'react';
import { EntityInfoDataTableGraphed, EntityInfoPanel } from '../../components';
import { isNodeResponse, isRelationshipResponse, useExploreSelectedItem, usePrimaryKind } from '../../hooks';
import { useIsEnterprise } from '../../providers/AppNameProvider';
import { EdgeInfoPane } from './EdgeInfo';

const defaultClasses: HTMLProps<HTMLElement>['className'] = 'bottom-0 top-0 py-4 absolute right-4';

const getItemKinds = (item: NodeDetails | RelationshipDetails | undefined) => {
    if (!item) return [];

    return isNodeResponse(item) ? item.kinds : [item.kind];
};

const GraphItemInformationPanel = () => {
    const { selectedItem, selectedItemQuery } = useExploreSelectedItem();

    const showFilteringBanner = useIsEnterprise();

    const kinds = getItemKinds(selectedItemQuery.data);
    const primaryKind = usePrimaryKind(kinds);

    if (!selectedItem || selectedItemQuery.isLoading) {
        return null;
    }

    if (selectedItemQuery.isError) {
        return (
            <EntityInfoPanel
                showFilteringBanner={showFilteringBanner}
                DataTable={EntityInfoDataTableGraphed}
                className={defaultClasses}
                selectedNode={{ graphId: selectedItem, id: '', name: 'Unknown', type: 'Unknown' }}
            />
        );
    }

    if (!selectedItemQuery.data) return null;

    if (isRelationshipResponse(selectedItemQuery.data)) {
        return <EdgeInfoPane className={defaultClasses} selectedEdge={selectedItemQuery.data} />;
    }

    if (isNodeResponse(selectedItemQuery.data)) {
        const selectedNode = {
            graphId: selectedItemQuery.data.node_id.toString(),
            id: selectedItemQuery.data.properties.objectid ?? '',
            name: selectedItemQuery.data.properties.name || selectedItemQuery.data.properties.objectid || '',
            type: primaryKind,
        };
        return (
            <EntityInfoPanel
                showFilteringBanner={showFilteringBanner}
                className={defaultClasses}
                selectedNode={selectedNode}
                DataTable={EntityInfoDataTableGraphed}
            />
        );
    }
};

export default GraphItemInformationPanel;
