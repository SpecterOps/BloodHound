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

import { Menu, MenuContent, MenuTrigger } from 'doodle-ui';
import { FC, ReactNode } from 'react';
import GraphButton from '../GraphButton';

const GraphMenu: FC<{
    label: string;
    children: ReactNode;
}> = ({ children, label }) => {
    return (
        <Menu>
            <MenuTrigger asChild>
                <GraphButton
                    aria-label={label}
                    data-testid={`explore_graph-controls_${label.toLowerCase().split(' ').join('-')}-menu`}
                    displayText={label}
                />
            </MenuTrigger>
            <MenuContent side='top'>{children}</MenuContent>
        </Menu>
    );
};

export default GraphMenu;
