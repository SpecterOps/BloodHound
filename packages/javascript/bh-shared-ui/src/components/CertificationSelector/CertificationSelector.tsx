// Copyright 2025 Specter Ops, Inc.
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

import { Button } from '@bloodhoundenterprise/doodleui';
import { Box, MenuItem, Popover } from '@mui/material';
import { FC, useState } from 'react';
import { cn } from '../../utils/theme';
import { AppIcon } from '../AppIcon';
import { CertificationOption } from './types';

const CertificationSelector: FC<{
    icon: any;
    options: CertificationOption[];
    selectedText: string;
    onChange: (selection: CertificationOption) => void;
}> = ({ icon, options, selectedText, onChange }) => {
    const [anchorEl, setAnchorEl] = useState(null);
    const open = Boolean(anchorEl);

    const handleClick = (e: any) => {
        setAnchorEl(e.currentTarget);
    };

    const handleClose = () => {
        setAnchorEl(null);
    };

    return (
        <Box p={1}>
            <Button
                variant='transparent'
                //className='truncate ring-1'
                onClick={handleClick}
                data-testid='certification-selector'>
                <span className='inline-flex justify-between gap-4 items-center w-full'>
                    <span>{icon}</span>
                    <span>{selectedText}</span>
                    <span className={cn({ 'rotate-180 transition-transform': open })}>
                        <AppIcon.CaretDown size={12} />
                    </span>
                </span>
            </Button>
            <Popover
                open={open}
                anchorEl={anchorEl}
                onClose={handleClose}
                anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'left',
                }}
                transformOrigin={{
                    vertical: -10,
                    horizontal: 0,
                }}
                style={{ width: 250 }}
                data-testid='certification-selector-popover'>
                {options.map((option) => {
                    return (
                        <MenuItem
                            style={{
                                display: 'flex',
                                justifyContent: 'space-between',
                                width: 450,
                                maxWidth: 450,
                            }}
                            key={option.key}
                            onClick={() => {
                                onChange(option);
                                handleClose();
                            }}>
                            {option.value}
                        </MenuItem>
                    );
                })}
            </Popover>
        </Box>
    );
};

export default CertificationSelector;
