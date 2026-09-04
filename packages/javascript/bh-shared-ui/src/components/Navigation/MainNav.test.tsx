// Copyright 2026 Specter Ops, Inc.
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

import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, within } from '../../test-utils';
import { AppIcon } from '../AppIcon';
import MainNav from './MainNav';
import { MainNavData, MainNavDataListItem, MainNavLogoDataObject } from './types';

const MainNavLogoData: MainNavLogoDataObject = {
    project: {
        route: '/',
        icon: <AppIcon.BHCELogo size={24} />,
    },
    specterOps: {
        image: {
            imageUrl: `/test`,
            dimensions: { height: 40, width: 165 },
            classes: 'ml-4',
            altText: 'BHE Text Logo',
        },
    },
};
const MainNavPrimaryListData: MainNavDataListItem[] = [
    {
        label: 'Link Item',
        icon: <AppIcon.LineChart size={24} />,
        route: '/test',
        testId: 'global_nav-test-link',
    },
    {
        label: 'Link Item 2',
        icon: <AppIcon.LineChart size={24} />,
        route: '/secondroute',
        testId: 'global_nav-test-link-2',
    },
];

const handleClick = vi.fn();

const MainNavSecondaryListData: MainNavDataListItem[] = [
    {
        label: 'Action Item',
        icon: <AppIcon.LineChart size={24} />,
        onClick: handleClick,
        testId: 'global_nav-test-action',
    },
];

const mainNavData: MainNavData = {
    logo: MainNavLogoData,
    primaryList: MainNavPrimaryListData,
    secondaryList: MainNavSecondaryListData,
};

const currentVersionNumber = 'v999.999.999';

const server = setupServer(
    rest.get(`/api/version`, async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    API: {
                        current_version: 'v2',
                        deprecated_version: 'v1',
                    },
                    server_version: currentVersionNumber,
                },
            })
        );
    })
);
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('MainNav', () => {
    const user = userEvent.setup();

    beforeEach(() => {
        render(<MainNav mainNavData={mainNavData} />);
    });
    it('should render a nav element with logo, two lists, a version number and a powered by', () => {
        expect(screen.getByRole('navigation')).toBeInTheDocument();
        expect(screen.getByTestId('global_nav-home')).toBeInTheDocument();
        expect(screen.getByTestId('global_nav-primary-list')).toBeInTheDocument();
        expect(screen.getByTestId('global_nav-secondary-list')).toBeInTheDocument();
        expect(screen.getByTestId('global_nav-version-number')).toBeInTheDocument();
        expect(screen.getByTestId('global_nav-powered-by')).toBeInTheDocument();
    });
    it('should render a navigation list item', async () => {
        const testLinkItem = MainNavPrimaryListData[0];

        const primaryList = await screen.findByTestId('global_nav-primary-list');
        const linkItem = await within(primaryList).getAllByTestId('global_nav-test-link')[0];
        const linkItemIcon = await within(primaryList).getAllByTestId('global_nav-item-label-icon')[0];
        const linkItemText = await within(primaryList).findByText(testLinkItem.label as string);

        expect(linkItem).toBeInTheDocument();
        expect(linkItem).toHaveAttribute('href', testLinkItem.route);
        expect(linkItemIcon).toBeInTheDocument();
        expect(linkItemText).toBeInTheDocument();
    });
    it('should render action list item that handles a function', async () => {
        const testLinkItem = MainNavSecondaryListData[0];

        const secondaryList = await screen.findByTestId('global_nav-secondary-list');
        const actionItem = await within(secondaryList).findByRole('button');
        const actionItemIcon = await within(secondaryList).findByTestId('global_nav-item-label-icon');
        const actionItemText = await within(secondaryList).findByText(testLinkItem.label as string);

        expect(actionItem).toBeInTheDocument();
        expect(actionItemIcon).toBeInTheDocument();
        expect(actionItemText).toBeInTheDocument();

        await user.click(actionItem);

        expect(testLinkItem.onClick).toBeCalled();
    });
    it('should render a label and version number when expanded', async () => {
        const MainNavBar = await screen.findByRole('navigation');

        const versionNumberContainer = await within(MainNavBar).findByTestId('global_nav-version-number');
        const versionNumberLabel = await within(versionNumberContainer).findByText(
            `BloodHound: ${currentVersionNumber}`
        );

        expect(versionNumberLabel).toBeInTheDocument();
    });
    it('should have a .z-nav class', () => {
        const navbarElement = screen.getByRole('navigation');
        expect(navbarElement).toHaveClass('z-nav');
    });
});

describe('Main Nav Route Highlighting', () => {
    it('should highlight selected route', () => {
        render(<MainNav mainNavData={mainNavData} />, {
            route: '/test',
        });
        expect(window.location.pathname).toBe('/test');
        const elem = screen.getByTestId('global_nav-test-link').closest('li');
        expect(elem).toHaveClass('bg-neutral-4');
        expect(screen.getByTestId('global_nav-test-link')).toHaveAttribute('aria-current', 'page');
    });
    it('should highlight main nav route when navigating to child route', () => {
        render(<MainNav mainNavData={mainNavData} />, {
            route: '/secondroute/child',
        });
        const selected = screen.getByTestId('global_nav-test-link-2').closest('li');
        const unselected = screen.getByTestId('global_nav-test-link').closest('li');
        expect(selected).toHaveClass('bg-neutral-4');
        expect(unselected).not.toHaveClass('bg-neutral-light-4');
    });
});

describe('Refreshed MainNav', () => {
    beforeEach(() => {
        localStorage.clear();
    });

    it('applies refreshed geometry, typography, and a semantic icon-only collapsed DOM', async () => {
        const user = userEvent.setup();
        render(<MainNav mainNavData={mainNavData} variant='refreshed' />);

        const navigation = screen.getByRole('navigation', { name: 'Global navigation' });
        expect(navigation).toHaveClass('basis-[264px]', 'w-[264px]', 'bg-primary-variant');
        expect(navigation).not.toHaveClass('shadow-md');
        const logo = screen.getByTestId('global_nav-home');
        expect(logo).toHaveClass('ml-2', 'mt-4', 'h-[30px]', 'w-[166px]');
        expect(logo.parentElement).toHaveClass('h-[72px]', 'border-b');
        const homeLink = screen.getByRole('link', { name: 'BloodHound home' });
        expect(homeLink).toHaveClass(
            'focus-visible:focus-ring-inset',
            'focus-visible:[--focus-ring:var(--common-white)]'
        );
        const toggle = screen.getByRole('button', { name: 'Toggle Navigation' });
        expect(toggle).toHaveClass(
            'h-8',
            'w-8',
            'left-[248px]',
            'bg-primary',
            'focus-visible:focus-ring-inset',
            'focus-visible:[--focus-ring:var(--common-white)]'
        );
        expect(screen.getByText('Link Item')).toHaveClass('typography-h5');
        expect(screen.getByText('Action Item')).toHaveClass('typography-body1');

        await user.click(toggle);

        expect(navigation).toHaveClass('basis-14', 'w-14');
        expect(screen.getByTestId('global_nav-home')).toHaveClass('w-10', 'h-[30px]');
        expect(homeLink).toHaveClass(
            'focus-visible:focus-ring-inset',
            'focus-visible:[--focus-ring:var(--common-white)]'
        );
        expect(screen.queryAllByTestId('global_nav-item-label-text')).toHaveLength(0);
        expect(screen.queryByTestId('global_nav-version-number')).not.toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Link Item' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Action Item' })).toBeInTheDocument();
    });
});

describe('Keyboard shortcuts', () => {
    it('should navigate to the correct page on alt + digit keydown', async () => {
        const user = userEvent.setup();
        render(<MainNav mainNavData={mainNavData} />);

        await user.keyboard('{Alt>}1{/Alt}');

        expect(window.location.pathname).toBe('/test');

        await user.keyboard('{Alt>}2{/Alt}');

        expect(window.location.pathname).toBe('/secondroute');

        await user.keyboard('{Alt>}1{/Alt}');

        expect(window.location.pathname).toBe('/test');
    });
});
