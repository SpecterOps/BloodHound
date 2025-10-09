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
import { Popover, PopoverContent, PopoverTrigger } from 'doodle-ui';
import { FC, MouseEvent } from 'react';
import { AppIcon } from '../../../../components';
interface SaveQueryActionMenuProps {
    saveAs: () => void;
}

const SaveQueryActionMenu: FC<SaveQueryActionMenuProps> = ({ saveAs }) => {
    const handleSaveAs = (event: MouseEvent) => {
        event.stopPropagation();
        saveAs();
    };

    const listItemStyles = 'px-2 py-3 cursor-pointer hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-4';

    return (
        <Popover>
            <PopoverTrigger
                className='inline-flex items-center justify-center whitespace-nowrap rounded-3xl text-sm ring-offset-background transition-colors hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 active:no-underline bg-neutral-light-5 text-neutral-dark-0 shadow-outer-1 hover:bg-secondary hover:text-white h-9 px-4 py-1 text-xs rounded-l-none pl-2 -ml-1 dark:text-neutral-dark-1 dark:hover:text-white'
                onClick={(event) => event.stopPropagation()}>
                <AppIcon.CaretDown size={10} />
            </PopoverTrigger>
            <PopoverContent className='p-0 w-28'>
                <div className={listItemStyles} onClick={handleSaveAs}>
                    Save As
                </div>
            </PopoverContent>
        </Popover>
    );
};

export default SaveQueryActionMenu;
