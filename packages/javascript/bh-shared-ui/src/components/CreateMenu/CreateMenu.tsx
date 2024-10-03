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
