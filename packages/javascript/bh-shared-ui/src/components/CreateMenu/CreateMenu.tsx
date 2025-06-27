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

import { Button } from '@bloodhoundenterprise/doodleui';
import { faCaretDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Menu, MenuItem, Typography } from '@mui/material';
import React, { FC } from 'react';
import { BaseSVGProps } from '../AppIcon/Icons/utils';
import FeatureFlag from '../FeatureFlag';

type MenuItems = { title: string; onClick: () => void; Icon?: FC<BaseSVGProps> }[];

const MenuWithDropdown: React.FC<{
    menuTitle: string;
    menuItems: MenuItems;
    CustomIcon?: FC<BaseSVGProps>;
    customStyles?: string;
    variant?: any;
    disabled: boolean;
}> = ({ menuTitle, menuItems, CustomIcon, customStyles, variant, disabled }) => {
    const buttonRef = React.useRef(null);
    const [isOpen, setIsOpen] = React.useState(false);

    const openMenu = () => {
        setIsOpen(true);
    };

    const closeMenu = () => {
        setIsOpen(false);
    };

    return (
        <>
            <Button
                aria-controls='create-menu'
                aria-haspopup='true'
                ref={buttonRef}
                variant={variant}
                className={customStyles}
                onClick={openMenu}
                disabled={disabled}>
                <Box display='flex' alignItems={'center'}>
                    {CustomIcon && <CustomIcon className='mr-2' />}
                    <Typography mr='8px'>{menuTitle}</Typography>
                    <FontAwesomeIcon icon={faCaretDown} />
                </Box>
            </Button>
            <Menu id='create-menu' anchorEl={buttonRef.current} keepMounted open={isOpen} onClose={closeMenu}>
                {menuItems.map((menuItem) => (
                    <MenuItem
                        key={menuItem.title}
                        onClick={() => {
                            menuItem.onClick();
                            closeMenu();
                        }}>
                        {menuItem.Icon && <menuItem.Icon className='mr-2' />}
                        {menuItem.title}
                    </MenuItem>
                ))}
            </Menu>
        </>
    );
};

const MenuOrButton: React.FC<{
    menuTitle: string;
    menuItems: MenuItems;
    CustomIcon?: FC<BaseSVGProps>;
    customStyles?: string;
    variant?: any;
    disabled: boolean;
}> = ({ menuTitle, menuItems, CustomIcon, customStyles, variant, disabled }) => {
    if (menuItems.length > 1) {
        return (
            <MenuWithDropdown
                menuItems={menuItems}
                menuTitle={menuTitle}
                disabled={disabled}
                CustomIcon={CustomIcon}
                customStyles={customStyles}
                variant={variant}
            />
        );
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
    CustomIcon?: FC<BaseSVGProps>;
    customStyles?: string;
    variant?: any;
    disabled?: boolean;
    featureFlag?: string;
    featureFlagEnabledMenuItems?: MenuItems;
}> = ({
    createMenuTitle,
    menuItems,
    CustomIcon,
    customStyles,
    variant,
    featureFlag,
    featureFlagEnabledMenuItems,
    disabled = false,
}) => {
    const menuOrButton = (
        <MenuOrButton
            menuTitle={createMenuTitle}
            menuItems={menuItems}
            CustomIcon={CustomIcon}
            customStyles={customStyles}
            variant={variant}
            disabled={disabled}
        />
    );

    if (featureFlag !== undefined && !!featureFlagEnabledMenuItems) {
        const featureFlagEnabledMenuOrButton = (
            <MenuOrButton
                menuTitle={createMenuTitle}
                menuItems={featureFlagEnabledMenuItems}
                CustomIcon={CustomIcon}
                customStyles={customStyles}
                variant={variant}
                disabled={disabled}
            />
        );

        return <FeatureFlag flagKey={featureFlag} enabled={featureFlagEnabledMenuOrButton} disabled={menuOrButton} />;
    } else {
        return menuOrButton;
    }
};

export default CreateMenu;
