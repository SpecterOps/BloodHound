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
import { faCheck, faXmark } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { EntityField, cn, format } from '../../../utils';
import { validateProperty } from '../../../utils/entityInfoDisplay';
import CopyToClipboardButton from '../../CopyToClipboardButton';
import NodeIcon from '../../NodeIcon';

const FALLBACK_STRING = '--';

const ExploreTableDataCell = ({ value, columnKey }: { value: EntityField['value']; columnKey: string }) => {
    if (columnKey === 'kind') {
        return (
            <div className='flex justify-center'>
                <NodeIcon nodeType={value?.toString() || ''} />
            </div>
        );
    }
    if (typeof value === 'boolean') {
        return (
            <div className='flex justify-center items-center pb-1 pt-1'>
                <FontAwesomeIcon
                    icon={value ? faCheck : faXmark}
                    className={cn(
                        'scale-125 fill-current',
                        value ? 'text-green-600' : 'text-gray-600 dark:text-gray-400'
                    )}
                />
            </div>
        );
    }

    const stringyKey = columnKey?.toString();
    const { kind } = validateProperty(columnKey);
    const formattedValue = format({ kind, keyprop: stringyKey, value, label: stringyKey });

    return formattedValue ? (
        <span className='cursor-auto'>
            <CopyToClipboardButton value={formattedValue} />
        </span>
    ) : (
        FALLBACK_STRING
    );
};

export default ExploreTableDataCell;
