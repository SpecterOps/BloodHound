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
import { Menu } from '@mui/material';
import { IconButton, Tooltip } from 'doodle-ui';
import React, { Children, FC, JSXElementConstructor, ReactElement, useState } from 'react';

type RenderableChild = ReactElement<any, string | JSXElementConstructor<any>>;
type Attributes = Partial<React.HTMLAttributes<Element>>;

const GraphMenu: FC<{
    label: string;
    icon: IconDefinition;
    tooltip?: string;
    children: RenderableChild | RenderableChild[];
}> = ({ children, label, icon, tooltip }) => {
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

    const open = Boolean(anchorEl);

    const handleClose = () => setAnchorEl(null);

    const testId = `explore_graph-controls_${label.toLowerCase().split(' ').join('-')}-menu`;
    const handleTriggerClick = (event: React.MouseEvent<HTMLButtonElement>) => {
        setAnchorEl(event.currentTarget);
    };

    const trigger = (
        <Tooltip
            tooltip={<span>{tooltip ?? label}</span>}
            triggerProps={{ className: 'pointer-events-auto' }}
            contentProps={{ className: 'dark:bg-neutral-4 dark:border-neutral-5 dark:text-white' }}>
            <div>
                <IconButton
                    aria-label={label}
                    data-testid={testId}
                    onClick={handleTriggerClick}
                    aria-controls={open ? `${label}-menu` : undefined}
                    aria-haspopup='true'
                    aria-expanded={open ? 'true' : undefined}>
                    <FontAwesomeIcon icon={icon} />
                </IconButton>
            </div>
        </Tooltip>
    );

    return (
        <>
            {trigger}
            <Menu
                id={`${label}-menu`}
                open={open}
                anchorEl={anchorEl}
                onClose={handleClose}
                MenuListProps={{
                    'aria-labelledby': `${label}-button`,
                }}
                anchorOrigin={{
                    vertical: 'top',
                    horizontal: 'left',
                }}
                transformOrigin={{
                    vertical: 'bottom',
                    horizontal: 'left',
                }}>
                {Children.map(children, (child) => {
                    if (React.isValidElement(child) && child.props && (child.props as Attributes)?.onClick) {
                        try {
                            return React.cloneElement(child, {
                                onClick: (e: React.MouseEvent) => {
                                    (child?.props as Attributes).onClick?.(e);
                                    handleClose();
                                },
                            } as Attributes);
                        } catch (e) {
                            return child;
                        }
                    }

                    return child;
                })}
            </Menu>
        </>
    );
};

export default GraphMenu;
