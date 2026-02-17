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
import { Tooltip } from '@bloodhoundenterprise/doodleui';
import { faAngleDoubleUp, faRemove } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import React from 'react';
import Icon from '../../components/Icon';
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
    const { clearSelectedItem, selectedItem, selectedItemType } = useExploreSelectedItem();

    const handleCollapseAll = () => {
        setIsObjectInfoPanelOpen(false);

        if (expandedPanelSections?.length) {
            setExploreParams({
                expandedPanelSections: [],
            });
        }
    };

    const hiddenNode = nodeType === 'HIDDEN' && selectedItemType === 'node';

    return (
        <div className='flex justify-between items-center text-sm font-bold pr-4'>
            {selectedItem ? (
                <Icon
                    className='h-10 box-border p-4 text-contrast'
                    onClick={clearSelectedItem}
                    tip='Clear selected item'>
                    <FontAwesomeIcon icon={faRemove} />
                </Icon>
            ) : (
                <div className='w-3' />
            )}

            {hiddenNode ? <HiddenEntityIcon /> : <NodeIcon nodeType={nodeType} />}

            <Tooltip tooltip={name} contentProps={{ side: 'bottom' }}>
                <h6
                    data-testid='explore_entity-information-panel_header-text'
                    className='truncate pr-2 leading-10 grow ml-2'>
                    {name}
                </h6>
            </Tooltip>

            <Icon
                tip='Collapse All'
                onClick={handleCollapseAll}
                className='box-border text-contrast'
                data-testid='explore_entity-information-panel_button-collapse-all'>
                <FontAwesomeIcon icon={faAngleDoubleUp} />
            </Icon>
        </div>
    );
};

export default Header;
