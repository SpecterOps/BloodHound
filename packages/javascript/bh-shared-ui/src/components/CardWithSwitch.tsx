// Copyright 2024 Specter Ops, Inc.
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

import { Switch } from 'doodle-ui';
import { FC, ReactNode } from 'react';
import { cn } from '../utils';

type CardWithSwitchProps = {
    title: string;
    description?: string;
    isEnabled: boolean;
    children?: ReactNode;
    disableSwitch?: boolean;
    onSwitchChange: () => void;
};

const CardWithSwitch: FC<CardWithSwitchProps> = ({
    title,
    description,
    isEnabled,
    onSwitchChange,
    children,
    disableSwitch = false,
}) => {
    return (
        <div
            className={cn('p-4 border rounded-lg', {
                'bg-neutral-2 border-transparent shadow-outer-1': isEnabled,
                'bg-neutral-2/30 border-neutral-3 shadow-none': !isEnabled,
            })}>
            <div className='flex justify-between mb-4'>
                <h4 className='font-bold text-lg'>{title}</h4>
                <Switch
                    label={isEnabled ? 'On' : 'Off'}
                    checked={isEnabled}
                    onCheckedChange={onSwitchChange}
                    disabled={disableSwitch}></Switch>
            </div>
            {children || <p>{description}</p>}
        </div>
    );
};

export default CardWithSwitch;
