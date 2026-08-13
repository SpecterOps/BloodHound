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
// TODO - move to DUI BED-7635
import { IconButton, Tooltip } from 'doodle-ui';
import { FC, MouseEvent, PropsWithChildren } from 'react';
import { cn } from '../utils';

interface Props {
    tip: string;
    onClick?: (event: MouseEvent) => void;
    className?: string;
}
// TODO BED-6062
const Icon: FC<PropsWithChildren<Props>> = ({ tip, onClick: click, children, className }): JSX.Element => {
    const icon = (
        <IconButton aria-label={tip} className={cn('relative rounded-none', className)} onClick={click}>
            {children}
        </IconButton>
    );

    return tip ? (
        <Tooltip tooltip={tip} contentProps={{ side: 'bottom', align: 'start' }}>
            {icon}
        </Tooltip>
    ) : (
        icon
    );
};

export default Icon;
