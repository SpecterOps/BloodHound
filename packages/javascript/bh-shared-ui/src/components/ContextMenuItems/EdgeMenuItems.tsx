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
import { getEdgeType } from '../../edgeTypes';
import { type PathfindingFilters } from '../../hooks';

type Props = {
    id: string;
    pathfindingFilters: PathfindingFilters;
};

export const EdgeMenuItems: FC<Props> = ({ id, pathfindingFilters }) => {
    const { handleUpdateAndApplyFilter } = pathfindingFilters;

    const edgeType = getEdgeType(id);

    const filterEdge = () => {
        if (edgeType) {
            handleUpdateAndApplyFilter(edgeType);
        }
    };

    // Prevent filtering for edge types not found in AllEdgeTypes array
    return edgeType ? (
        <MenuItem key='filter-edge' onClick={filterEdge}>
            Filter out Edge
        </MenuItem>
    ) : (
        <MenuItem key='non-filterable' disabled>
            Non-filterable Edge
        </MenuItem>
    );
};
