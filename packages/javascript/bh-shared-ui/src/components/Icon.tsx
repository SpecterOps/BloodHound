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

import { Button, Tooltip } from '@bloodhoundenterprise/doodleui';
import { FC, MouseEvent, PropsWithChildren } from 'react';
import { cn } from '../utils';

interface Props {
    tip?: string;
    onClick?: (event: MouseEvent) => void;
    badge?: number;
    className?: string;
}

const Icon: FC<PropsWithChildren<Props>> = ({ tip, onClick: click, children, badge = 0, className }): JSX.Element => {
    const overflow: boolean = badge > 99;
    const badgeText: string | null = overflow ? '99+' : badge > 0 ? badge.toString() : null;

    const icon = (
        <Button variant={'text'} className={cn('relative p-0 rounded-none', className)} onClick={click}>
            {children}
            {badgeText && <Badge text={badgeText} overflow={overflow} />}
        </Button>
    );

    return tip ? (
        <Tooltip tooltip={tip} contentProps={{ side: 'bottom' }}>
            {icon}
        </Tooltip>
    ) : (
        icon
    );
};

const Badge: FC<{ text: string; overflow?: boolean }> = ({ text, overflow = false }): JSX.Element => {
    return (
        <span
            className={cn('absolute bottom-[3px] right-[3px] size-5 text-xs rounded-lg leading-5', {
                'leading-[21px]': overflow,
            })}>
            {text}
        </span>
    );
};

export default Icon;
