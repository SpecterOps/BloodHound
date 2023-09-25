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

import { Menu } from '@mui/material';
import GraphButton from '../GraphButton';
import { Children, FC, ReactNode, useState } from 'react';

const GraphMenu: FC<{ label: string; children: ReactNode }> = ({ children, label }) => {
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

    const open = Boolean(anchorEl);

    const handleClose = () => setAnchorEl(null);

    return (
        <>
            <GraphButton
                onClick={(event: React.MouseEvent<HTMLButtonElement>) => {
                    setAnchorEl(event.currentTarget);
                }}
                aria-controls={open ? `${label}-menu` : undefined}
                aria-haspopup='true'
                aria-expanded={open ? 'true' : undefined}
                displayText={label}></GraphButton>
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
                    return <div onClick={handleClose}>{child}</div>;
                })}
            </Menu>
        </>
    );
};

export default GraphMenu;
