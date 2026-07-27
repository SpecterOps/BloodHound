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
    EntityInfoDataTableGraphed,
    EntityInfoPanel,
    GraphItemInformationPanel as SharedGraphItemInformationPanel,
    isNodeResponse,
    useExploreSelectedItem,
} from 'bh-shared-ui';
import { HTMLProps } from 'react';
import { getRACFAdditionalTables } from 'src/racfhound/racfAdditionalTables';

const defaultClasses: HTMLProps<HTMLElement>['className'] = 'bottom-0 top-0 py-4 absolute right-4';

// GraphItemInformationPanel is an app-local wrapper around the shared (upstream) panel.
//
// Upstream moved GraphItemInformationPanel into bh-shared-ui, which cannot import our
// app-local RACF table components. So for RACF node kinds we render EntityInfoPanel directly
// and inject the RACF relationship tables. Every other case (edges, errors, loading, and
// non-RACF nodes) delegates to the shared panel so we inherit its behavior without
// duplicating it. This keeps all RACF customization in the app layer and touches no
// shared-library files, minimizing conflicts on future upstream merges.
//
// The tables are injected as `priorityTables`, not `additionalTables`: EntityInfoContent only
// renders `additionalTables` for built-in kinds (via EntityInfoList), routing custom/OpenGraph
// kinds like RACF to KindInfoItems, which ignores them. `priorityTables` render unconditionally
// (above the object information), so RACF sections show for these custom kinds.
const GraphItemInformationPanel = () => {
    const { selectedItem, selectedItemQuery } = useExploreSelectedItem();
    const data = selectedItemQuery.data;

    if (selectedItem && data && isNodeResponse(data)) {
        const racfTables = getRACFAdditionalTables(data, selectedItem);

        if (racfTables) {
            return (
                <EntityInfoPanel
                    className={defaultClasses}
                    selectedNode={data}
                    DataTable={EntityInfoDataTableGraphed}
                    priorityTables={racfTables}
                />
            );
        }
    }

    return <SharedGraphItemInformationPanel />;
};

export default GraphItemInformationPanel;
