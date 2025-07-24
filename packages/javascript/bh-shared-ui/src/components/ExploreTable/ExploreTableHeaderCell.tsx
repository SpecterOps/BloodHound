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
import { faCaretDown, faCaretUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Tooltip } from '@mui/material';
import { cn, formatPotentiallyUnknownLabel } from '../../utils';
import { MungedTableRowWithId } from './explore-table-utils';

const KEYS_TO_RENDER_AS_ICON = ['nodetype'];

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
    return (
        <Tooltip title={<p>{label}</p>}>
            <div
                className={cn(
                    'flex items-center m-0 cursor-pointer h-full w-full hover:bg-neutral-100 dark:hover:bg-neutral-dark-4',
                    {
                        'justify-center':
                            dataType === 'boolean' || KEYS_TO_RENDER_AS_ICON.includes(headerKey.toString()),
                    }
                )}
                onClick={onClick}>
                <div className='truncate'>{label}</div>
                <div className={cn('pl-2', { ['opacity-0']: sortBy !== headerKey })}>
                    {!sortOrder && <FontAwesomeIcon icon={faCaretDown} />}
                    {sortOrder === 'asc' && <FontAwesomeIcon icon={faCaretUp} />}
                    {sortOrder === 'desc' && <FontAwesomeIcon icon={faCaretDown} />}
                </div>
            </div>
        </Tooltip>
    );
};

export default ExploreTableHeaderCell;
