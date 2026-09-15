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
import { IconButton } from 'doodle-ui';
import React from 'react';
import HiddenEntityIcon from '../../../components/HiddenEntityIcon';
import { useExploreParams, useExploreSelectedItem } from '../../../hooks';
import { useObjectInfoPanelContext } from '../providers';

export interface HeaderProps {
    name: string;
}

const Header: React.FC<HeaderProps> = ({ name = 'None Selected' }) => {
    const { setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();
    const { setExploreParams } = useExploreParams();
    const { clearSelectedItem, isHidden } = useExploreSelectedItem();

    const handleCollapseAll = () => {
        setIsObjectInfoPanelOpen(false);
        setExploreParams({
            expandedPanelSections: [],
        });
    };

    return (
        <div className='flex justify-between items-center text-sm font-bold mx-2 gap-2'>
            <IconButton
                aria-label='Collapse All'
                onClick={handleCollapseAll}
                data-testid='explore_edge-information-pane_button-collapse-all'>
                <FontAwesomeIcon icon={faAngleDoubleUp} />
            </IconButton>

            {isHidden && <HiddenEntityIcon />}

            <h2 data-testid='explore_edge-information-pane_header-text' className='truncate pl-2 pr-4 leading-10 grow'>
                {name}
            </h2>

            <IconButton aria-label='Clear selected item' onClick={clearSelectedItem}>
                <FontAwesomeIcon icon={faRemove} />
            </IconButton>
        </div>
    );
};

export default Header;
