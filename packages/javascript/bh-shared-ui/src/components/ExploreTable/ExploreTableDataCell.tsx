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
import { faCancel, faCheck } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { EntityField, format } from '../../utils';
import NodeIcon from '../NodeIcon';
import { isIconField } from './explore-table-utils';

const FALLBACK_STRING = '--';

const ExploreTableDataCell = ({ value, columnKey }: { value: EntityField['value']; columnKey: string }) => {
    if (columnKey === 'nodetype') {
        return (
            <div className='flex justify-center'>
                <NodeIcon nodeType={value?.toString() || ''} />
            </div>
        );
    }
    if (isIconField(value)) {
        return (
            <div className='flex justify-center items-center pb-1 pt-1'>
                <FontAwesomeIcon
                    icon={value ? faCheck : faCancel}
                    color={value ? 'green' : 'lightgray'}
                    className='scale-125'
                />
            </div>
        );
    }

    const stringyKey = columnKey?.toString();

    return (
        format({ keyprop: stringyKey, value, label: stringyKey }) || (
            <div className='w-full text-center'>{FALLBACK_STRING}</div>
        )
    );
};

export default ExploreTableDataCell;
