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

import { Box } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import clsx from 'clsx';
import React, { DataHTMLAttributes } from 'react';

const useStyles = makeStyles((theme) => ({
    container: {
        display: 'flex',
        flexDirection: 'row',
        alignItems: 'center',
        borderBottom: '3px solid transparent',
        cursor: 'pointer',
        marginRight: theme.spacing(4),
        whiteSpace: 'nowrap',
        '&:hover': {
            borderBottom: `3px solid #a7adb0`,
        },
        '& svg': {
            color: '#a7adb0',
        },
    },
    icon: {
        marginRight: theme.spacing(1),
    },
    title: {
        textTransform: 'uppercase',
        fontSize: '0.875rem',
        lineHeight: 1.5,
        fontWeight: 500,
        letterSpacing: '0.0075em',
    },
    active: {
        color: '#406f8e',
        borderBottom: `3px solid #6798B9`,
        '&:hover': {
            borderBottom: `3px solid #6798B9`,
        },
        '& svg': {
            color: '#6798B9',
        },
    },
}));

interface MenuItemProps extends DataHTMLAttributes<HTMLDivElement> {
    title: string;
    active: boolean;
    icon?: React.ReactElement;
    onClick: () => void;
}

const MenuItem: React.FC<MenuItemProps> = ({ title, active, icon, onClick, ...rest }) => {
    const classes = useStyles();

    return (
        <div className={clsx(classes.container, active ? classes.active : null)} onClick={onClick} {...rest}>
            {icon && <Box className={classes.icon}>{icon}</Box>}
            <Box className={clsx(classes.title, 'noselect')}>{title}</Box>
        </div>
    );
};

export default MenuItem;
