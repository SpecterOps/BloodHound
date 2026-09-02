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
import { usePrimaryKind } from '../../hooks';
import { EntityInfoContentProps, EntityKinds, allSections } from '../../utils';

const getEntityKindInfoSections = (id: string, kind: string) => {
    if (kind === ActiveDirectoryNodeKind.LocalGroup || kind === ActiveDirectoryNodeKind.LocalUser)
        return allSections[ActiveDirectoryNodeKind.Entity]!(id);

    if (allSections[kind as EntityKinds]) return allSections[kind as EntityKinds]!(id);

    return [];
};

const EntityInfoList: React.FC<EntityInfoContentProps> = ({ selectedNode, additionalTables, DataTable }) => {
    const primaryKind = usePrimaryKind(selectedNode.kinds);
    const entityInfoKindSections = getEntityKindInfoSections(selectedNode.properties.objectid ?? '', primaryKind);

    return (
        <div data-testid='entity-info-data-table-list'>
            {entityInfoKindSections.map((table, index) => {
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

export default EntityInfoList;
