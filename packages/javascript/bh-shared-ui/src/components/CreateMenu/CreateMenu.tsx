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

import { faCaretDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Typography } from '@mui/material';
import { Button, Menu, MenuContent, MenuItem, MenuTrigger } from 'doodle-ui';
import React, { ComponentPropsWithoutRef, FC } from 'react';
import FeatureFlag from '../FeatureFlag';

type MenuItems = { title: string; onClick: () => void }[];

const MenuWithDropdown: React.FC<{
    menuTitle: string;
    menuItems: MenuItems;
    disabled: boolean;
}> = ({ menuTitle, menuItems, disabled }) => {
    return (
        <Menu>
            <MenuTrigger asChild>
                <Button disabled={disabled}>
                    <Box display='flex' alignItems='center'>
                        <Typography mr='8px'>{menuTitle}</Typography>
                        <FontAwesomeIcon icon={faCaretDown} />
                    </Box>
                </Button>
            </MenuTrigger>
            <MenuContent>
                {menuItems.map((menuItem) => (
                    <MenuItem key={menuItem.title} onSelect={menuItem.onClick}>
                        {menuItem.title}
                    </MenuItem>
                ))}
            </MenuContent>
        </Menu>
    );
};

const MenuOrButton: React.FC<{
    menuTitle: string;
    menuItems: MenuItems;
    disabled: boolean;
}> = ({ menuTitle, menuItems, disabled }) => {
    if (menuItems.length > 1) {
        return <MenuWithDropdown menuItems={menuItems} menuTitle={menuTitle} disabled={disabled} />;
    } else if (menuItems.length === 1) {
        return (
            <Button
                disabled={disabled}
                onClick={() => {
                    menuItems[0].onClick();
                }}>
                {menuItems[0].title}
            </Button>
        );
    }
    return null;
};

const CreateMenu: React.FC<{
    createMenuTitle: string;
    menuItems: MenuItems;
    CustomIcon?: FC<{ className: string }>;
    customStyles?: string;
    variant?: ComponentPropsWithoutRef<typeof Button>['variant'];
    disabled?: boolean;
    featureFlag?: string;
    featureFlagEnabledMenuItems?: MenuItems;
}> = ({ createMenuTitle, menuItems, featureFlag, featureFlagEnabledMenuItems, disabled = false }) => {
    const menuOrButton = <MenuOrButton menuTitle={createMenuTitle} menuItems={menuItems} disabled={disabled} />;

    if (featureFlag !== undefined && !!featureFlagEnabledMenuItems) {
        const featureFlagEnabledMenuOrButton = (
            <MenuOrButton menuTitle={createMenuTitle} menuItems={featureFlagEnabledMenuItems} disabled={disabled} />
        );

        return <FeatureFlag flagKey={featureFlag} enabled={featureFlagEnabledMenuOrButton} disabled={menuOrButton} />;
    } else {
        return menuOrButton;
    }
};

export default CreateMenu;
