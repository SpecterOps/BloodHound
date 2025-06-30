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
import { Box, Divider } from '@mui/material';
import React from 'react';
import { ActiveDirectoryNodeKind } from '../../graphSchema';
import { EntityKinds, allSections } from '../../utils';
import EntityInfoDataTable from './EntityInfoDataTable';
import EntityInfoDataTableGraphed from './EntityInfoDataTableGraphed';
import EntitySelectorsInformation from './EntitySelectorsInformation';

export interface EntityInfoContentProps {
    id: string;
    nodeType: EntityKinds | string;
    databaseId?: string;
    additionalSections?: boolean;
}

const EntityInfoDataTableList: React.FC<EntityInfoContentProps> = ({ id, nodeType, additionalSections }) => {
    let type = nodeType as EntityKinds;
    if (nodeType === ActiveDirectoryNodeKind.LocalGroup || nodeType === ActiveDirectoryNodeKind.LocalUser)
        type = ActiveDirectoryNodeKind.Entity;

    const tables = allSections[type]?.(id) || [];

    if (additionalSections) {
        tables.push({ id, label: 'Selectors' });
    }

    return (
        <div data-testid='entity-info-data-table-list'>
            {tables.map((table, index) => {
                if (table.label === 'Selectors') {
                    return <EntitySelectorsInformation key='selectors' />;
                }

                const TableComponent = additionalSections ? EntityInfoDataTable : EntityInfoDataTableGraphed;
                return (
                    <React.Fragment key={index}>
                        <Box padding={1}>
                            <Divider />
                        </Box>
                        <TableComponent {...table} />
                    </React.Fragment>
                );
            })}
        </div>
    );
};

export default EntityInfoDataTableList;
