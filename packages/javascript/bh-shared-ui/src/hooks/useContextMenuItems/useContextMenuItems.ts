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

import { UseQueryResult } from 'react-query';
import { MousePosition } from '../../types';
import { Permission } from '../../utils';
import { useExploreParams } from '../useExploreParams';
import { useExploreSelectedItem } from '../useExploreSelectedItem';
import { isEdge, isNode, type ItemResponse } from '../useGraphItem';
import { usePermissions } from '../usePermissions';

type QueryResponse = UseQueryResult<ItemResponse, unknown>;

const NAV_MENU_WIDTH = 56;

/** Return position to show context menu, with nav menu offset */
const getMenuPosition = (position: MousePosition | null) =>
    position
        ? {
              left: position.mouseX + NAV_MENU_WIDTH,
              top: position.mouseY,
          }
        : null;

export const useContextMenuItems = (position: MousePosition | null) => {
    const { selectedItemQuery } = useExploreSelectedItem();
    const exploreParams = useExploreParams();
    const { exploreSearchTab } = exploreParams;
    const { checkPermission } = usePermissions();

    /** Checks and returns the data in a query typed as an EdgeResponse */
    const asEdgeItem = (query: QueryResponse) =>
        isEdge(query.data) && exploreSearchTab === 'pathfinding' ? query.data : undefined;

    /** Checks and returns the data in a query typed as an NodeResponse */
    const asNodeItem = (query: QueryResponse) => (isNode(query.data) ? query.data : undefined);

    const isAssetGroupEnabled = checkPermission(Permission.GRAPH_DB_WRITE);

    return {
        asEdgeItem,
        asNodeItem,
        exploreParams,
        isAssetGroupEnabled,
        menuPosition: getMenuPosition(position),
        selectedItemQuery,
    };
};
