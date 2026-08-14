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
import { NodeDetails, NodeDetailsWithInfo } from 'js-client-library';
import React, { HTMLProps } from 'react';
import { usePrimaryKind } from '../../hooks';
import { EntityInfoDataTableProps, NoEntitySelectedMessage, cn, getEntityName } from '../../utils';
import { ObjectInfoPanelContextProvider } from '../../views/Explore/providers/ObjectInfoPanelProvider';
import { RoleBasedFilterBadge } from '../RoleBasedFilterBadge';
import EntityInfoContent from './EntityInfoContent';
import Header from './EntityInfoHeader';

type EntityTable = React.FC<EntityInfoDataTableProps>;

export type EntityTables = {
    sectionProps: EntityInfoDataTableProps;
    TableComponent: EntityTable;
}[];

export interface EntityInfoPanelProps {
    DataTable: EntityTable;
    selectedNode?: NodeDetails | NodeDetailsWithInfo;
    className?: HTMLProps<HTMLDivElement>['className'];
    additionalTables?: EntityTables;
    priorityTables?: EntityTables;
    showPlaceholderMessage?: boolean;
}

const EntityInfoPanel: React.FC<EntityInfoPanelProps> = ({
    selectedNode,
    className,
    additionalTables,
    priorityTables,
    DataTable,
    showPlaceholderMessage = false,
}) => {
    const primaryKind = usePrimaryKind(selectedNode?.kinds ?? []);
    return (
        <div
            className={cn(
                'flex flex-col rounded-lg pointer-events-none overflow-y-hidden h-full min-w-[400px] w-[400px] max-w-[400px] gap-2',
                className
            )}
            data-testid='explore_entity-information-panel'>
            <RoleBasedFilterBadge />
            <div className='bg-neutral-2 pointer-events-auto rounded-lg shadow-outer-1'>
                <Header name={getEntityName(selectedNode)} nodeType={primaryKind} />
            </div>
            <div className='bg-neutral-2 overflow-x-hidden overflow-y-auto py-1 px-4 pointer-events-auto rounded-lg shadow-outer-1'>
                {selectedNode ? (
                    <EntityInfoContent
                        DataTable={DataTable}
                        priorityTables={priorityTables}
                        additionalTables={additionalTables}
                        selectedNode={selectedNode}
                    />
                ) : (
                    <p className='text-sm'>
                        {showPlaceholderMessage
                            ? 'Select an object to view the associated information'
                            : NoEntitySelectedMessage}
                    </p>
                )}
            </div>
        </div>
    );
};

const WrappedEntityInfoPanel: React.FC<EntityInfoPanelProps> = (props) => (
    <ObjectInfoPanelContextProvider>
        <EntityInfoPanel {...props} />
    </ObjectInfoPanelContextProvider>
);

export default WrappedEntityInfoPanel;
