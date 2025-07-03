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
import { EntityInfoDataTableProps, EntityKinds, allSections } from '../../utils';

export interface EntityInfoContentProps {
    DataTable: React.FC<EntityInfoDataTableProps>;
    id: string;
    nodeType: EntityKinds | string;
    databaseId?: string;
    additionalTables?: [
        {
            sectionProps: EntityInfoDataTableProps;
            TableComponent: React.FC<EntityInfoDataTableProps>;
        },
    ];
}

const EntityInfoDataTableList: React.FC<EntityInfoContentProps> = ({ id, nodeType, additionalTables, DataTable }) => {
    let type = nodeType as EntityKinds;
    if (nodeType === ActiveDirectoryNodeKind.LocalGroup || nodeType === ActiveDirectoryNodeKind.LocalUser)
        type = ActiveDirectoryNodeKind.Entity;

    const tables = allSections[type]?.(id) || [];

    return (
        <div data-testid='entity-info-data-table-list'>
            {tables.map((table, index) => {
                return (
                    <React.Fragment key={index}>
                        <Box padding={1}>
                            <Divider />
                        </Box>
                        <DataTable {...table} />
                    </React.Fragment>
                );
            })}

            {additionalTables?.map((table, index) => {
                const { sectionProps, TableComponent } = table;
                return (
                    <React.Fragment key={index}>
                        <Box padding={1}>
                            <Divider />
                        </Box>
                        <TableComponent {...sectionProps} />
                    </React.Fragment>
                );
            })}
        </div>
    );
};

export default EntityInfoDataTableList;
