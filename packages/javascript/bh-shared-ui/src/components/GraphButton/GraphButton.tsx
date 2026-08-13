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

import { Button, ButtonProps } from 'doodle-ui';
import { FC, ReactNode } from 'react';
import { cn } from '../../utils';

export interface GraphButtonProps extends Omit<ButtonProps, 'children'> {
    displayText: string | ReactNode;
}

const GraphButton: FC<GraphButtonProps> = (props) => {
    const { displayText, className, ...attributes } = props;

    return (
        <Button
            {...attributes}
            className={cn(
                'box-content h-4 min-w-0 rounded p-3 text-base font-medium leading-4 capitalize',
                'bg-[#F4F4F4] text-[#1D1B20] dark:bg-[#222222] dark:text-white',
                'hover:bg-[#E3E7EA] hover:no-underline dark:hover:bg-[#272727]',
                className
            )}>
            {displayText}
        </Button>
    );
};

export default GraphButton;
