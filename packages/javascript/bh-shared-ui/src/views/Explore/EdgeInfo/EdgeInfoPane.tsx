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
import useRoleBasedFiltering from '../../../hooks/useRoleBasedFiltering';
import { SelectedEdge } from '../../../store';
import { cn } from '../../../utils';
import { ObjectInfoPanelContextProvider } from '../providers';
import EdgeInfoContent from './EdgeInfoContent';
import Header from './EdgeInfoHeader';

interface EdgeInfoPaneProps {
    selectedEdge: SelectedEdge | null;
    className?: HTMLProps<HTMLDivElement>['className'];
}

const EdgeInfoPane: React.FC<EdgeInfoPaneProps> = ({ className, selectedEdge }) => {
    const isRoleBasedFiltering = useRoleBasedFiltering();

    return (
        <div
            className={cn(
                'flex flex-col pointer-events-none overflow-y-hidden h-full w-[400px] max-w-[400px]',
                className
            )}
            data-testid='explore_edge-information-pane'>
            {isRoleBasedFiltering && (
                <Badge
                    data-testid='explore_entity-information-panel-badge-etac-filtering'
                    className='justify-start text-sm text-neutral-dark-1 bg-[#F8EEFD] dark:bg-[#472E54] dark:text-neutral-light-1 border-0 mb-2'
                    icon={<FontAwesomeIcon icon={faEyeSlash} className='mr-2' />}
                    label='Role-based access filtering applied'
                />
            )}
            <div className='bg-neutral-2 pointer-events-auto rounded'>
                <Header name={selectedEdge?.name || 'None'} />
            </div>
            <div className='bg-neutral-2 mt-2 overflow-x-hidden overflow-y-auto py-1 px-4 pointer-events-auto rounded'>
                {selectedEdge === null ? 'No information to display.' : <EdgeInfoContent selectedEdge={selectedEdge} />}
            </div>
        </div>
    );
};

const WrappedEdgeInfoPane: React.FC<EdgeInfoPaneProps> = (props) => (
    <ObjectInfoPanelContextProvider>
        <EdgeInfoPane {...props} />
    </ObjectInfoPanelContextProvider>
);

export default WrappedEdgeInfoPane;
