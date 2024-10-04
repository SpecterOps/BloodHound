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
import React from 'react';
import FeatureFlag from '../FeatureFlag';

type MenuItems = { title: string; onClick: () => void }[];

const MenuWithDropdown: React.FC<{ menuTitle: string; menuItems: MenuItems }> = ({ menuTitle, menuItems }) => {
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
            <Button aria-controls='create-menu' aria-haspopup='true' ref={buttonRef} onClick={openMenu}>
                <Box display='flex' alignItems={'center'}>
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
                        {menuItem.title}
                    </MenuItem>
                ))}
            </Menu>
        </>
    );
};

const CreateMenu: React.FC<{
    createMenuTitle: string;
    menuItems: MenuItems;
    featureFlag?: string;
    featureFlagEnabledMenuItems?: MenuItems;
}> = ({ createMenuTitle, menuItems, featureFlag, featureFlagEnabledMenuItems }) => {
    const menuOrButton =
        menuItems.length > 1 ? (
            <MenuWithDropdown menuTitle={createMenuTitle} menuItems={menuItems} />
        ) : (
            <Button
                onClick={() => {
                    menuItems[0].onClick();
                }}>
                {createMenuTitle}
            </Button>
        );

    if (featureFlag != undefined && !!featureFlagEnabledMenuItems) {
        const featureFlagEnabledMenuOrButton =
            featureFlagEnabledMenuItems.length > 1 ? (
                <MenuWithDropdown menuTitle={createMenuTitle} menuItems={featureFlagEnabledMenuItems} />
            ) : (
                <Button
                    onClick={() => {
                        featureFlagEnabledMenuItems[0].onClick();
                    }}>
                    {featureFlagEnabledMenuItems[0].title}
                </Button>
            );
        return <FeatureFlag flagKey={featureFlag} enabled={featureFlagEnabledMenuOrButton} disabled={menuOrButton} />;
    } else {
        return menuOrButton;
    }
};

export default CreateMenu;
