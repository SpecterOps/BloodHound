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
import { cn, formatPotentiallyUnknownLabel } from '../../utils';
import { MungedTableRowWithId } from './explore-table-utils';

const ExploreTableHeaderCell = ({
    headerKey,
    sortBy,
    sortOrder,
    onClick,
}: {
    headerKey: keyof MungedTableRowWithId;
    sortBy?: keyof MungedTableRowWithId;
    sortOrder?: string;
    onClick: () => void;
}) => {
    return (
        <div
            className='flex items-center p-1 m-0 cursor-pointer h-full hover:bg-neutral-100 dark:hover:bg-neutral-dark-4'
            onClick={onClick}>
            <div>{formatPotentiallyUnknownLabel(String(headerKey))}</div>
            <div className={cn('pl-2', { ['opacity-0']: sortBy !== headerKey })}>
                {!sortOrder && <FontAwesomeIcon icon={faCaretDown} />}
                {sortOrder === 'asc' && <FontAwesomeIcon icon={faCaretUp} />}
                {sortOrder === 'desc' && <FontAwesomeIcon icon={faCaretDown} />}
            </div>
        </div>
    );
};

export default ExploreTableHeaderCell;
