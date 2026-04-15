import type { Meta, StoryObj } from '@storybook/react';
import { MagnifyingGlass } from '../../styleguide/components/AppIcons/components/MagnifyingGlass';
import { User } from '../../styleguide/components/AppIcons/components/User';
import { UserCog } from '../../styleguide/components/AppIcons/components/UserCog';
import {
    Menu,
    MenuContent,
    MenuItem,
    MenuLabel,
    MenuSeparator,
    MenuSub,
    MenuSubContent,
    MenuSubTrigger,
    MenuTrigger,
} from './Menu';

const meta = {
    title: 'Components/Menu',
    component: MenuItem,
    tags: ['autodocs'],
    parameters: { layout: 'centered' },
    argTypes: {
        icon: {
            description: 'Icon element rendered on the left side of the item. Requires `iconLeft` to be true.',
            control: false,
        },
        iconLeft: {
            description: 'Whether to show the icon on the left side of the item.',
            control: 'boolean',
        },
        secondaryMenu: {
            description:
                'Renders a caret-right indicator on the right edge of the item. Use `MenuSub` + `MenuSubTrigger` for functional submenus.',
            control: 'boolean',
        },
        disabled: {
            description: 'Makes the item non-interactive and applies disabled styling.',
            control: 'boolean',
        },
        children: {
            description: 'The label content of the menu item.',
            control: 'text',
        },
        onSelect: {
            description: 'Callback fired when the item is selected.',
        },
    },
    args: {
        children: 'Menu Item',
        iconLeft: false,
        secondaryMenu: false,
        disabled: false,
    },
} satisfies Meta<typeof MenuItem>;

export default meta;

type Story = StoryObj<typeof meta>;

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
                <MenuItem onSelect={() => console.log('Profile')}>Profile</MenuItem>
                <MenuSub>
                    <MenuSubTrigger>Settings</MenuSubTrigger>
                    <MenuSubContent>
                        <MenuItem onSelect={() => console.log('Account')}>Account</MenuItem>
                        <MenuItem onSelect={() => console.log('Privacy')}>Privacy</MenuItem>
                        <MenuItem onSelect={() => console.log('Notifications')}>Notifications</MenuItem>
                    </MenuSubContent>
                </MenuSub>
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
                <MenuSub>
                    <MenuSubTrigger>
                        <UserCog size={16} className='mr-2' />
                        Settings Submenu
                    </MenuSubTrigger>
                    <MenuSubContent>
                        <MenuItem onSelect={() => console.log('Account')}>Account</MenuItem>
                        <MenuItem onSelect={() => console.log('Privacy')}>Privacy</MenuItem>
                    </MenuSubContent>
                </MenuSub>
                <MenuSeparator />
                <MenuItem icon={<MagnifyingGlass size={16} />} iconLeft disabled>
                    Disabled with Icon
                </MenuItem>
            </MenuContent>
        </Menu>
    ),
};
