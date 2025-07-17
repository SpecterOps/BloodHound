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

import { MenuItem } from '@mui/material';
import { FC } from 'react';
import { useExploreParams } from '../../hooks';

type Props = {
    exploreParams: ReturnType<typeof useExploreParams>;
    objectId: string;
};

export const NodeMenuItems: FC<Props> = ({ exploreParams, objectId }) => {
    const { primarySearch, secondarySearch, setExploreParams } = exploreParams;

    return (
        <>
            <MenuItem
                key='starting-node'
                onClick={() =>
                    setExploreParams({
                        exploreSearchTab: 'pathfinding',
                        searchType: secondarySearch ? 'pathfinding' : 'node',
                        primarySearch: objectId,
                    })
                }>
                Set as starting node
            </MenuItem>

            <MenuItem
                key='ending-node'
                onClick={() =>
                    setExploreParams({
                        exploreSearchTab: 'pathfinding',
                        searchType: primarySearch ? 'pathfinding' : 'node',
                        secondarySearch: objectId,
                    })
                }>
                Set as ending node
            </MenuItem>
        </>
    );
};
