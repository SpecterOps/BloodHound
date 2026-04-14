import type { Meta, StoryObj } from '@storybook/react';
import { MagnifyingGlass } from '../../styleguide/components/AppIcons/components/MagnifyingGlass';
import { User } from '../../styleguide/components/AppIcons/components/User';
import { UserCog } from '../../styleguide/components/AppIcons/components/UserCog';
import { Menu, MenuContent, MenuItem, MenuLabel, MenuSeparator, MenuTrigger } from './Menu';

const meta: Meta<typeof Menu> = {
    title: 'Components/Menu',
    component: Menu,
};

export default meta;

type Story = StoryObj<typeof Menu>;

const Trigger = () => (
    <MenuTrigger asChild>
        <button style={{ padding: '8px 12px', border: '1px solid #ccc', borderRadius: 4 }}>Open Menu</button>
    </MenuTrigger>
);

export const Default: Story = {
    render: () => (
        <Menu>
            <Trigger />
            <MenuContent>
                <MenuLabel>My Account</MenuLabel>
                <MenuSeparator />
                <MenuItem onSelect={() => console.log('Profile')}>Profile</MenuItem>
                <MenuItem onSelect={() => console.log('Settings')}>Settings</MenuItem>
                <MenuSeparator />
                <MenuItem onSelect={() => console.log('Logout')}>Logout</MenuItem>
            </MenuContent>
        </Menu>
    ),
};

export const WithIconLeft: Story = {
    render: () => (
        <Menu>
            <Trigger />
            <MenuContent>
                <MenuLabel>My Account</MenuLabel>
                <MenuSeparator />
                <MenuItem icon={<User size={16} />} iconLeft onSelect={() => console.log('Profile')}>
                    Profile
                </MenuItem>
                <MenuItem icon={<UserCog size={16} />} iconLeft onSelect={() => console.log('Settings')}>
                    Settings
                </MenuItem>
                <MenuSeparator />
                <MenuItem icon={<MagnifyingGlass size={16} />} iconLeft onSelect={() => console.log('Search')}>
                    Search
                </MenuItem>
            </MenuContent>
        </Menu>
    ),
};

export const WithSecondaryMenu: Story = {
    render: () => (
        <Menu>
            <Trigger />
            <MenuContent>
                <MenuLabel>Navigation</MenuLabel>
                <MenuSeparator />
                <MenuItem secondaryMenu onSelect={() => console.log('Profile')}>
                    Profile
                </MenuItem>
                <MenuItem secondaryMenu onSelect={() => console.log('Settings')}>
                    Settings
                </MenuItem>
            </MenuContent>
        </Menu>
    ),
};

export const WithDisabledItem: Story = {
    render: () => (
        <Menu>
            <Trigger />
            <MenuContent>
                <MenuLabel>My Account</MenuLabel>
                <MenuSeparator />
                <MenuItem onSelect={() => console.log('Profile')}>Profile</MenuItem>
                <MenuItem disabled onSelect={() => console.log('Settings')}>
                    Settings (disabled)
                </MenuItem>
                <MenuSeparator />
                <MenuItem onSelect={() => console.log('Logout')}>Logout</MenuItem>
            </MenuContent>
        </Menu>
    ),
};

export const AllFeatures: Story = {
    render: () => (
        <Menu>
            <Trigger />
            <MenuContent>
                <MenuLabel>All Features</MenuLabel>
                <MenuSeparator />
                <MenuItem icon={<User size={16} />} iconLeft onSelect={() => console.log('Profile')}>
                    With Icon Left
                </MenuItem>
                <MenuItem icon={<UserCog size={16} />} iconLeft secondaryMenu onSelect={() => console.log('Settings')}>
                    Icon Left + Secondary Menu
                </MenuItem>
                <MenuItem secondaryMenu onSelect={() => console.log('More')}>
                    Secondary Menu Only
                </MenuItem>
                <MenuSeparator />
                <MenuItem icon={<MagnifyingGlass size={16} />} iconLeft disabled>
                    Disabled with Icon
                </MenuItem>
            </MenuContent>
        </Menu>
    ),
};
