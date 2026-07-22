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

import { Menu, Tooltip } from '@mui/material';
import React, { Children, FC, ReactNode, useState } from 'react';
import GraphButton from '../GraphButton';

type Attributes = Partial<React.HTMLAttributes<Element>>;

const GraphMenu: FC<{
    controlId: string;
    label: string;
    displayText?: ReactNode;
    children: ReactNode;
}> = ({ children, controlId, displayText, label }) => {
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

    const open = Boolean(anchorEl);
    const buttonId = `graph-${controlId}-button`;
    const menuId = `graph-${controlId}-menu`;

    const handleClose = () => setAnchorEl(null);

    const menuButton = (
        <GraphButton
            id={buttonId}
            aria-label={label}
            data-testid={`explore_graph-controls_${controlId}-menu`}
            onClick={(event: React.MouseEvent<HTMLButtonElement>) => {
                setAnchorEl(event.currentTarget);
            }}
            aria-controls={menuId}
            aria-haspopup='menu'
            aria-expanded={open}
            displayText={displayText ?? label}
        />
    );

    return (
        <>
            {displayText ? (
                <Tooltip placement='top' title={label}>
                    {menuButton}
                </Tooltip>
            ) : (
                menuButton
            )}
            <Menu
                open={open}
                anchorEl={anchorEl}
                onClose={handleClose}
                keepMounted
                MenuListProps={{
                    id: menuId,
                    'aria-labelledby': buttonId,
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
