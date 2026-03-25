// Copyright 2026 Specter Ops, Inc.
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

import { ButtonProps, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import DropdownTriggerContents from './DropdownTriggerContents';

const DropdownTrigger: FC<{
    open: boolean;
    selectedText: JSX.Element | string;
    buttonProps?: ButtonProps;
    StartAdornment?: React.FC;
    EndAdornment?: React.FC;
    testId?: string;
    variant?: ButtonProps['variant'];
}> = (props) => {
    return (
        <PopoverTrigger asChild>
            <DropdownTriggerContents {...props} />
        </PopoverTrigger>
    );
};

export default DropdownTrigger;
