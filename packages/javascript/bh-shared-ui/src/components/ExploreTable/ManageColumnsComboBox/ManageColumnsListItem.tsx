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

import { Checkbox } from '@bloodhoundenterprise/doodleui';
import { faThumbTack } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { UseComboboxPropGetters, useMultipleSelection } from 'downshift';
import { ManageColumnsComboBoxOption } from './ManageColumnsComboBox';

type ManageColumnsListItemProps = {
    isSelected?: boolean;
    item: ManageColumnsComboBoxOption;
    onClick:
        | ReturnType<typeof useMultipleSelection<ManageColumnsComboBoxOption>>['removeSelectedItem']
        | ReturnType<typeof useMultipleSelection<ManageColumnsComboBoxOption>>['addSelectedItem'];
    itemProps: ReturnType<UseComboboxPropGetters<ManageColumnsComboBoxOption>['getItemProps']>;
};

const ManageColumnsListItem = ({ isSelected, item, onClick, itemProps }: ManageColumnsListItemProps) => (
    <li
        className={`p-2 w-full hover:bg-gray-100 ${isSelected ? 'cursor-default' : 'cursor-pointer'} ${item.isPinned ? 'bg-gray-100' : ''}`}
        {...itemProps}
        aria-disabled={item?.isPinned}
        onClick={(e) => {
            e.stopPropagation();
            onClick(item);
        }}>
        <button
            className={`w-full text-left flex justify-between items-center ${isSelected ? 'cursor-default' : 'cursor-pointer'}`}>
            <div>
                <Checkbox className={`mr-2 ${isSelected ? '*:bg-blue-800' : ''}`} checked={isSelected} />
                <span>{item.value}</span>
            </div>
            {item.isPinned && (
                <FontAwesomeIcon
                    fill='white'
                    stroke=''
                    className='justify-self-end stroke-cyan-300'
                    icon={faThumbTack}
                />
            )}
        </button>
    </li>
);

export default ManageColumnsListItem;
