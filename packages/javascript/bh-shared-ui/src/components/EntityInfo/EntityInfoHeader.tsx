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
import { faAngleDoubleUp, faRemove } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton, Tooltip } from 'doodle-ui';
import React from 'react';
import NodeIcon from '../../components/NodeIcon/NodeIcon';
import { useExploreParams, useExploreSelectedItem } from '../../hooks';
import { EntityKinds } from '../../utils/content';
import { useObjectInfoPanelContext } from '../../views/Explore/providers';
import HiddenEntityIcon from '../HiddenEntityIcon';

export interface HeaderProps {
    name: string;
    nodeType?: EntityKinds | string;
}
const Header: React.FC<HeaderProps> = ({ name, nodeType }) => {
    const { setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();
    const { setExploreParams, expandedPanelSections } = useExploreParams();
    const { clearSelectedItem, selectedItem, isHidden } = useExploreSelectedItem();

    const handleCollapseAll = () => {
        setIsObjectInfoPanelOpen(false);

        if (expandedPanelSections?.length) {
            setExploreParams({
                expandedPanelSections: [],
            });
        }
    };

    return (
        <div className='flex justify-between items-center text-sm font-bold'>
            <IconButton
                aria-label='Collapse All'
                className='px-4'
                onClick={handleCollapseAll}
                data-testid='explore_entity-information-panel_button-collapse-all'>
                <FontAwesomeIcon icon={faAngleDoubleUp} />
            </IconButton>
            {isHidden ? <HiddenEntityIcon /> : <NodeIcon nodeType={nodeType} />}
            <Tooltip tooltip={name} contentProps={{ side: 'bottom' }}>
                <h6
                    data-testid='explore_entity-information-panel_header-text'
                    className='truncate pl-2 pr-4 leading-10 grow'>
                    {name}
                </h6>
            </Tooltip>
            {selectedItem && (
                <IconButton aria-label='Clear selected item' onClick={clearSelectedItem}>
                    <FontAwesomeIcon icon={faRemove} />
                </IconButton>
            )}
        </div>
    );
};

export default Header;
