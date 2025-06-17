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
import { Box } from '@mui/material';
import React from 'react';
import { useParams } from 'react-router-dom';
import { EntityKinds } from '../../utils';
import { SelectorsInfoPanelContextProvider } from '../../views/ZoneManagement/providers';
import EntityInfoDataTableList from './EntityInfoDataTableList';
import EntityObjectInformation from './EntityObjectInformation';
import EntitySelectorsInformation from './EntitySelectorstInformation';

export interface EntityInfoContentProps {
    id: string;
    nodeType: EntityKinds | string;
    databaseId?: string;
    zoneManagement?: boolean;
}

const EntityInfoContent: React.FC<EntityInfoContentProps> = (props, zoneManagement) => {
    const { tierId, labelId, memberId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;

    return (
        <Box>
            <EntityObjectInformation {...props} />
            <EntityInfoDataTableList {...props} />
            {zoneManagement && tagId !== undefined && memberId !== undefined && (
                <SelectorsInfoPanelContextProvider>
                    <EntitySelectorsInformation tagId={tagId} memberId={memberId} />
                </SelectorsInfoPanelContextProvider>
            )}
        </Box>
    );
};

export default EntityInfoContent;
