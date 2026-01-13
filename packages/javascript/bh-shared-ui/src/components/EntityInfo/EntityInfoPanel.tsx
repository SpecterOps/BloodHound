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
import { Badge } from '@bloodhoundenterprise/doodleui';
import { faEyeSlash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import React, { HTMLProps } from 'react';
import { useSelf } from '../../hooks/useBloodHoundUsers';
import { useListDisplayRoles } from '../../hooks/useListDisplayRoles/useListDisplayRoles';
import { privilegeZonesPath } from '../../routes';
import { SelectedNode } from '../../types';
import { EntityInfoDataTableProps, NoEntitySelectedHeader, NoEntitySelectedMessage, cn } from '../../utils';
import { isETACRole } from '../../utils/roles';
import { ObjectInfoPanelContextProvider } from '../../views/Explore/providers/ObjectInfoPanelProvider';
import EntityInfoContent from './EntityInfoContent';
import Header from './EntityInfoHeader';

export type EntityTables = {
    sectionProps: EntityInfoDataTableProps;
    TableComponent: React.FC<EntityInfoDataTableProps>;
}[];

interface EntityInfoPanelProps {
    DataTable: React.FC<EntityInfoDataTableProps>;
    selectedNode?: SelectedNode | null;
    className?: HTMLProps<HTMLDivElement>['className'];
    additionalTables?: EntityTables;
    priorityTables?: EntityTables;
}

const EntityInfoPanel: React.FC<EntityInfoPanelProps> = ({
    selectedNode,
    className,
    additionalTables,
    priorityTables,
    DataTable,
}) => {
    const isPrivilegeZonesPage = location.pathname.includes(`/${privilegeZonesPath}`);

    const getSelfQuery = useSelf();
    const getRolesQuery = useListDisplayRoles();
    const roles = getRolesQuery.data;
    const userRoleId = getSelfQuery?.data?.roles.map((item: any) => item.id);
    const selectedETACEnabledRole = isETACRole(Number(userRoleId), roles);
    const roleBasedFiltering: boolean = getSelfQuery?.data?.all_environments === false && selectedETACEnabledRole;

    return (
        <div
            className={cn(
                'flex flex-col pointer-events-none overflow-y-hidden h-full min-w-[400px] w-[400px] max-w-[400px]',
                className
            )}
            data-testid='explore_entity-information-panel'>
            {!isPrivilegeZonesPage && roleBasedFiltering && (
                <Badge
                    data-testid='explore_entity-information-panel-badge-etac-filtering'
                    className='justify-start text-sm text-neutral-dark-1 bg-[#F8EEFD] dark:bg-[#472E54] dark:text-neutral-light-1 border-0 mb-2'
                    icon={<FontAwesomeIcon icon={faEyeSlash} />}
                    label='&nbsp; Role-based access filtering applied'
                />
            )}
            <div className='bg-neutral-2 pointer-events-auto rounded'>
                <Header
                    name={selectedNode?.name ? selectedNode?.name : NoEntitySelectedHeader}
                    nodeType={selectedNode?.type}
                />
            </div>
            <div className='bg-neutral-2 mt-2 overflow-x-hidden overflow-y-auto py-1 px-4 pointer-events-auto rounded'>
                {selectedNode ? (
                    <EntityInfoContent
                        DataTable={DataTable}
                        id={selectedNode.id}
                        nodeType={selectedNode.type}
                        databaseId={selectedNode.graphId}
                        priorityTables={priorityTables}
                        additionalTables={additionalTables}
                    />
                ) : (
                    <p className='text-sm'>
                        {isPrivilegeZonesPage
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
