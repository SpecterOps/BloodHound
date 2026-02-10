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
import React from 'react';
import HiddenEntityIcon from '../../../components/HiddenEntityIcon';
import Icon from '../../../components/Icon';
import { useExploreParams, useExploreSelectedItem } from '../../../hooks';
import { useObjectInfoPanelContext } from '../providers';

export interface HeaderProps {
    name: string;
}

const Header: React.FC<HeaderProps> = ({ name = 'None Selected' }) => {
    const { setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();
    const { setExploreParams } = useExploreParams();
    const { clearSelectedItem, selectedItem, selectedItemType } = useExploreSelectedItem();

    const handleCollapseAll = () => {
        setIsObjectInfoPanelOpen(false);
        setExploreParams({
            expandedPanelSections: [],
        });
    };

    const hiddenEdge = selectedItem?.includes('HIDDEN') && selectedItemType === 'edge';

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

            {hiddenEdge && <HiddenEntityIcon />}

            <h6 data-testid='explore_edge-information-pane_header-text' className='text-nowrap leading-10 grow ml-2'>
                {name}
            </h6>

            <Icon
                tip='Collapse All'
                onClick={handleCollapseAll}
                className='h-10 box-border text-contrast'
                data-testid='explore_edge-information-pane_button-collapse-all'>
                <FontAwesomeIcon icon={faAngleDoubleUp} />
            </Icon>
        </div>
    );
};

export default Header;
