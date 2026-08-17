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
import { Button, Popover, PopoverContent, PopoverTrigger } from 'doodle-ui';
import { FC, KeyboardEvent, MouseEvent } from 'react';
import { AppIcon } from '../../../../components';
import { adaptClickHandlerToKeyDown } from '../../../../utils/adaptClickHandlerToKeyDown';
interface SaveQueryActionMenuProps {
    saveAs: () => void;
}

const SaveQueryActionMenu: FC<SaveQueryActionMenuProps> = ({ saveAs }) => {
    const handleSaveAs = <T extends MouseEvent | KeyboardEvent>(event: T) => {
        event.stopPropagation();
        saveAs();
    };

    const listItemStyles =
        'px-2 py-3 cursor-pointer hover:bg-dropdown-option-hover-fill focus-visible:bg-dropdown-option-hover-fill focus-visible:shadow-[inset_3px_0_0_var(--focus-ring)]';

    return (
        <Popover>
            <PopoverTrigger asChild aria-label='Show save query options' onClick={(event) => event.stopPropagation()}>
                <Button variant='secondary' size='small' className='rounded-l-none pl-2 -ml-1'>
                    <AppIcon.CaretDown size={10} />
                </Button>
            </PopoverTrigger>
            <PopoverContent className='p-0 w-28'>
                <div
                    role='button'
                    tabIndex={0}
                    onKeyDown={adaptClickHandlerToKeyDown(handleSaveAs)}
                    className={listItemStyles}
                    onClick={handleSaveAs}>
                    Save As
                </div>
            </PopoverContent>
        </Popover>
    );
};

export default SaveQueryActionMenu;
