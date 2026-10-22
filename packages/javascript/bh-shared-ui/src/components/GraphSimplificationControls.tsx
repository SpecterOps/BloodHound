// Copyright 2023 Specter Ops, Inc.
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

import { faCompressArrowsAlt, faExpandArrowsAlt } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton } from 'doodle-ui';

export function GraphSimplificationControls({
    canCollapse,
    canExpand,
    onCollapse,
    onExpand,
    collapseLabel = 'Collapse nodes',
    expandLabel = 'Expand all nodes',
}: {
    canCollapse: boolean;
    canExpand: boolean;
    onCollapse: () => void;
    onExpand: () => void;
    collapseLabel?: string;
    expandLabel?: string;
}) {
    return (
        <div className='flex gap-1 pointer-events-auto' role='group' aria-label='Graph simplification'>
            <IconButton aria-label={collapseLabel} disabled={!canCollapse} onClick={onCollapse}>
                <FontAwesomeIcon aria-hidden='true' icon={faCompressArrowsAlt} />
            </IconButton>
            <IconButton aria-label={expandLabel} disabled={!canExpand} onClick={onExpand}>
                <FontAwesomeIcon aria-hidden='true' icon={faExpandArrowsAlt} />
            </IconButton>
        </div>
    );
}
