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
import { EntityInfoContentProps } from '../../utils';

const EntityInfoDataTablePriorityList: React.FC<Pick<EntityInfoContentProps, 'priorityTables'>> = ({
    priorityTables,
}) => {
    if (!priorityTables || priorityTables.length == 0) {
        return null;
    }

    return (
        <div data-testid='entity-info-data-table-priority-list'>
            {priorityTables?.map((table, index) => {
                const { sectionProps, TableComponent } = table;
                return (
                    <React.Fragment key={index}>
                        <TableComponent {...sectionProps} />
                        <Box padding={1}>
                            <Divider />
                        </Box>
                    </React.Fragment>
                );
            })}
        </div>
    );
};

export default EntityInfoDataTablePriorityList;
