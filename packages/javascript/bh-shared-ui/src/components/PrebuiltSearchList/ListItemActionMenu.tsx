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
import { FC, MouseEvent } from 'react';

import { Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { useSavedQueriesContext } from '../../views/Explore/providers/SavedQueriesProvider';
import { VerticalEllipsis } from '../AppIcon/Icons';
interface ListItemActionMenuProps {
    id?: number;
    query?: string;
    deleteQuery: (id: number) => void;
}

const ListItemActionMenu: FC<ListItemActionMenuProps> = ({ id, query, deleteQuery }) => {
    const { runQuery, editQuery } = useSavedQueriesContext();

    const handleRun = (event: MouseEvent) => {
        event.stopPropagation();
        if (typeof query !== 'string') return;
        runQuery(query, id);
    };

    const handleEdit = (event: MouseEvent) => {
        event.stopPropagation();
        if (id == null) return;
        editQuery(id);
    };

    const handleDelete = (event: MouseEvent) => {
        event.stopPropagation();
        if (id == null) return;
        deleteQuery(id);
    };

    const listItemStyles = 'w-full px-2 py-3 cursor-pointer hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-4';

    return (
        <>
            <Popover>
                <PopoverTrigger
                    data-testid='saved-query-action-menu-trigger'
                    className='dark:text-white p-2 rounded rounded-full hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-2'
                    onClick={(event) => event.stopPropagation()}>
                    <VerticalEllipsis size={24} />
                </PopoverTrigger>
                <PopoverContent className='p-0' data-testid='saved-query-action-menu'>
                    <div className={listItemStyles} onClick={handleRun}>
                        Run
                    </div>
                    <div className={listItemStyles} onClick={handleEdit}>
                        Edit/Share
                    </div>
                    <div className={listItemStyles} onClick={handleDelete}>
                        Delete
                    </div>
                </PopoverContent>
            </Popover>
        </>
    );
};

export default ListItemActionMenu;
