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

import { FC, ReactNode } from 'react';
import { Paper, Box, Typography, useTheme } from '@mui/material';
import { addOpacityToHex } from '../utils/colors';
import { Switch } from '@bloodhoundenterprise/doodleui';

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
    const theme = useTheme();

    const enabledStyles = {
        background: theme.palette.background.paper,
        borderColor: 'transparent',
    };
    const disabledStyles = {
        background: addOpacityToHex(theme.palette.background.paper, 30),
        borderColor: theme.palette.neutral.tertiary,
        boxShadow: 'none',
    };

    return (
        <Paper
            sx={{
                padding: 2,
                borderWidth: '1px',
                borderStyle: 'solid',
                borderRadius: '8px',
                ...(isEnabled ? enabledStyles : disabledStyles),
            }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', marginBottom: '16px' }}>
                <Typography variant='h4'>{title}</Typography>
                <Switch
                    label={isEnabled ? 'On' : 'Off'}
                    checked={isEnabled}
                    onCheckedChange={onSwitchChange}
                    disabled={disableSwitch}></Switch>
            </Box>
            {children || <Typography>{description}</Typography>}
        </Paper>
    );
};

export default CardWithSwitch;
