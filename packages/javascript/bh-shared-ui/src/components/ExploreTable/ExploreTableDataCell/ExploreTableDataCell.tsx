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
import { faCancel, faCheck, faCopy } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useState } from 'react';
import { EntityField, cn, copyToClipboard, format } from '../../../utils';
import NodeIcon from '../../NodeIcon';

const FALLBACK_STRING = '--';

const transitionDelay = 'delay-300';

const ExploreTableDataCell = ({ value, columnKey }: { value: EntityField['value']; columnKey: string }) => {
    const [displayCopyCheckmark, setDisplayCopyCheckmark] = useState(false);
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
                    icon={value ? faCheck : faCancel}
                    color={value ? 'green' : 'lightgray'}
                    className='scale-125'
                />
            </div>
        );
    }

    const stringyKey = columnKey?.toString();
    const formattedValue = format({ keyprop: stringyKey, value, label: stringyKey });

    const handleCopyToClipBoard: React.MouseEventHandler<HTMLButtonElement> = (e) => {
        e.stopPropagation(); // prevents the click event bubbling up the DOM and triggering the row click handler
        if (typeof formattedValue === 'string') {
            copyToClipboard(formattedValue);
        } else {
            copyToClipboard(formattedValue.join(', '));
        }
        setDisplayCopyCheckmark(true);
        setTimeout(() => setDisplayCopyCheckmark(false), 500);
    };

    return formattedValue ? (
        <span className='cursor-auto'>
            <button
                onClick={handleCopyToClipBoard}
                className={cn(
                    'absolute top-1/2 left-2 -translate-x-1/2 -translate-y-1/2 opacity-0 pr-1 group-hover:opacity-100 transition-opacity ease-in',
                    transitionDelay
                )}>
                <FontAwesomeIcon icon={displayCopyCheckmark ? faCheck : faCopy} />
            </button>
            <span className={cn('group-hover:pl-5 transition-[padding-left] ease-in', transitionDelay)}>
                {formattedValue}
            </span>
        </span>
    ) : (
        FALLBACK_STRING
    );
};

export default ExploreTableDataCell;
