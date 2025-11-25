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
    className?: string;
}

const Icon: FC<PropsWithChildren<Props>> = ({ tip, onClick: click, children, className }): JSX.Element => {
    const icon = (
        <Button
            variant={'text'}
            className={cn('relative p-0 rounded-none', className)}
            onClick={click}
            aria-label={tip}>
            {children}
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

export default Icon;
