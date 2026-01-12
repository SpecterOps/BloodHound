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
import { Tooltip } from '@mui/material';
import { cn, formatPotentiallyUnknownLabel } from '../../utils';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';
import { AppIcon } from '../AppIcon';
import { MungedTableRowWithId } from './explore-table-utils';

const KEYS_TO_RENDER_AS_ICON = ['kind'];

const ExploreTableHeaderCell = ({
    headerKey,
    sortBy,
    sortOrder,
    onClick,
    dataType,
}: {
    headerKey: keyof MungedTableRowWithId;
    sortBy?: keyof MungedTableRowWithId;
    dataType: string;
    sortOrder?: string;
    onClick: () => void;
}) => {
    const label = formatPotentiallyUnknownLabel(String(headerKey));

    let IconComponent = AppIcon.SortEmpty;
    if (sortBy !== headerKey) {
        IconComponent = AppIcon.SortEmpty;
    } else {
        if (sortOrder === 'asc') IconComponent = AppIcon.SortAsc;
        if (sortOrder === 'desc') IconComponent = AppIcon.SortDesc;
    }

    return (
        <Tooltip title={<p>{label}</p>}>
            <div
                role='button'
                tabIndex={0}
                className={cn(
                    'flex items-center m-0 cursor-pointer h-full w-full hover:bg-neutral-100 dark:hover:bg-neutral-dark-4',
                    {
                        'justify-center':
                            dataType === 'boolean' || KEYS_TO_RENDER_AS_ICON.includes(headerKey.toString()),
                    }
                )}
                onClick={onClick}
                onKeyDown={adaptClickHandlerToKeyDown(onClick)}>
                <div className='truncate'>{label}</div>
                <div className='pl-2'>
                    <IconComponent size={12} />
                </div>
            </div>
        </Tooltip>
    );
};

export default ExploreTableHeaderCell;
