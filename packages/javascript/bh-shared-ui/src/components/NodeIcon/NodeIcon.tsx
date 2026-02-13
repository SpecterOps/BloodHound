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

import { Tooltip } from '@bloodhoundenterprise/doodleui';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useCustomNodeKinds } from '../../hooks/useCustomNodeKinds';
import { EntityKinds, MetaDetailNodeKind, MetaNodeKind } from '../../utils/content';
import { GetIconInfo } from '../../utils/icons';

interface NodeIconProps {
    nodeType?: EntityKinds | string;
}

function NodeIcon({ nodeType = '' }: NodeIconProps) {
    const customIcons = useCustomNodeKinds().data ?? {};
    const iconInfo = GetIconInfo(nodeType, customIcons);

    return (
        <Tooltip
            tooltip={nodeType}
            contentProps={{ className: 'bg-neutral-5 border-none text-contrast dark:text-contrast' }}>
            <div className='inline-block relative mr-1'>
                <div
                    className='flex items-center justify-center border border-neutral-dark-1 rounded-full p-1 size-[22px] text-neutral-dark-1 bg-neutral-light-1 pointer-events-none'
                    style={iconInfo ? { backgroundColor: iconInfo.color } : undefined}
                    title={nodeType}>
                    {nodeType === MetaNodeKind || nodeType === MetaDetailNodeKind ? (
                        <img src={'/ui/meta.png'} alt='meta node' className='size-full' />
                    ) : (
                        <FontAwesomeIcon icon={iconInfo.icon} transform='shrink-2' fixedWidth />
                    )}
                </div>
            </div>
        </Tooltip>
    );
}

export default NodeIcon;
