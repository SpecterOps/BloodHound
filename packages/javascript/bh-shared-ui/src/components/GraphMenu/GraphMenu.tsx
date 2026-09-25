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

import { IconDefinition } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton, Menu, MenuContent, MenuTrigger } from 'doodle-ui';
import { FC, ReactNode } from 'react';

const GraphMenu: FC<{
    label: string;
    icon: IconDefinition;
    children: ReactNode;
}> = ({ children, label, icon }) => {
    const testId = `explore_graph-controls_${label.toLowerCase().split(' ').join('-')}-menu`;

    return (
        <Menu>
            <MenuTrigger asChild>
                <IconButton aria-label={label} data-testid={testId}>
                    <FontAwesomeIcon icon={icon} />
                </IconButton>
            </MenuTrigger>
            <MenuContent side='top' align='start'>
                {children}
            </MenuContent>
        </Menu>
    );
};

export default GraphMenu;
