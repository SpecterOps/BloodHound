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
import { NAV_EXPANDED_STORAGE_KEY } from '../../hooks';
import { act, render, screen, waitFor, within } from '../../test-utils';
import { createMatchMediaController } from '../../testing';
import { SubNavItem } from '../../types';
import { AppIcon } from '../AppIcon';
import MainNav from './MainNav';
import { MainNavData, MainNavDataListItem, MainNavLogoDataObject, NavLinkItem } from './types';

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
const MainNavPrimaryListData: NavLinkItem[] = [
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
    it('should label the logo link as the BloodHound home link', () => {
        const homeLink = screen.getByRole('link', { name: 'BloodHound home' });
        expect(homeLink).toBeInTheDocument();
        expect(screen.getByTestId('global_nav-home')).toContainElement(homeLink);
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

describe('MainNav responsive expansion', () => {
    afterEach(() => {
        window.localStorage.clear();
        vi.unstubAllGlobals();
    });

    it('contracts an expanded nav when the viewport shrinks below the xl breakpoint', async () => {
        const matchMediaController = createMatchMediaController(true);
        window.localStorage.setItem(NAV_EXPANDED_STORAGE_KEY, JSON.stringify(true));

        render(<MainNav mainNavData={mainNavData} />);

        expect(screen.getByRole('button', { name: 'Toggle Navigation' })).toHaveAttribute('aria-expanded', 'true');

        act(() => matchMediaController.setMatches(false));

        await waitFor(() => {
            expect(screen.getByRole('button', { name: 'Toggle Navigation' })).toHaveAttribute('aria-expanded', 'false');
            expect(window.localStorage.getItem(NAV_EXPANDED_STORAGE_KEY)).toBe(JSON.stringify(false));
        });
    });

    it('expands a contracted nav when the viewport grows to the xl breakpoint', async () => {
        const matchMediaController = createMatchMediaController(false);
        window.localStorage.setItem(NAV_EXPANDED_STORAGE_KEY, JSON.stringify(false));

        render(<MainNav mainNavData={mainNavData} />);

        expect(screen.getByRole('button', { name: 'Toggle Navigation' })).toHaveAttribute('aria-expanded', 'false');

        act(() => matchMediaController.setMatches(true));

        await waitFor(() => {
            expect(screen.getByRole('button', { name: 'Toggle Navigation' })).toHaveAttribute('aria-expanded', 'true');
            expect(window.localStorage.getItem(NAV_EXPANDED_STORAGE_KEY)).toBe(JSON.stringify(true));
        });
    });

    it('preserves an expanded nav on initial load below the xl breakpoint', () => {
        createMatchMediaController(false);
        window.localStorage.setItem(NAV_EXPANDED_STORAGE_KEY, JSON.stringify(true));

        render(<MainNav mainNavData={mainNavData} />);

        expect(screen.getByRole('button', { name: 'Toggle Navigation' })).toHaveAttribute('aria-expanded', 'true');
        expect(window.localStorage.getItem(NAV_EXPANDED_STORAGE_KEY)).toBe(JSON.stringify(true));
    });

    it('preserves a contracted nav on initial load at the xl breakpoint', () => {
        createMatchMediaController(true);
        window.localStorage.setItem(NAV_EXPANDED_STORAGE_KEY, JSON.stringify(false));

        render(<MainNav mainNavData={mainNavData} />);

        expect(screen.getByRole('button', { name: 'Toggle Navigation' })).toHaveAttribute('aria-expanded', 'false');
        expect(window.localStorage.getItem(NAV_EXPANDED_STORAGE_KEY)).toBe(JSON.stringify(false));
    });

    it('preserves the user-selected state after a refresh without a breakpoint crossing', async () => {
        const user = userEvent.setup();
        createMatchMediaController(false);
        const { unmount } = render(<MainNav mainNavData={mainNavData} />);

        await user.click(screen.getByRole('button', { name: 'Toggle Navigation' }));

        expect(window.localStorage.getItem(NAV_EXPANDED_STORAGE_KEY)).toBe(JSON.stringify(false));

        unmount();
        render(<MainNav mainNavData={mainNavData} />);

        expect(screen.getByRole('button', { name: 'Toggle Navigation' })).toHaveAttribute('aria-expanded', 'false');
    });
});

describe('Main Nav Route Highlighting', () => {
    it('should highlight selected route', () => {
        render(<MainNav mainNavData={mainNavData} />, {
            route: '/test',
        });
        expect(window.location.pathname).toBe('/test');
        const elem = screen.getByTestId('global_nav-test-link').closest('li');
        expect(elem).toHaveClass('bg-primary');
        expect(screen.getByTestId('global_nav-test-link')).toHaveAttribute('aria-current', 'page');
    });
    it.each([
        ['/', '/', true],
        ['/test', '/test', true],
        ['/test', '/test/', true],
        ['/test', '/test/child', true],
        ['/test/*', '/test', true],
        ['/test/*', '/test/child', true],
        ['/test/', '/test', true],
        ['/test/', '/test/child', true],
        ['/test', '/test-other', false],
        ['/test', '/other/test', false],
        ['/test/*', '/test-other', false],
        ['/', '/test', false],
    ])('matches link %s at %s: %s', (route, pathname, active) => {
        render(
            <MainNav
                mainNavData={{
                    ...mainNavData,
                    primaryList: [{ ...MainNavPrimaryListData[0], route }],
                }}
            />,
            { route: pathname }
        );
        const link = screen.getByTestId('global_nav-test-link');
        expect(link.getAttribute('aria-current')).toBe(active ? 'page' : null);
        expect(link.closest('li')?.classList.contains('bg-primary')).toBe(active);
    });
    it('should highlight main nav route when navigating to child route', () => {
        render(<MainNav mainNavData={mainNavData} />, {
            route: '/secondroute/child',
        });
        const selected = screen.getByTestId('global_nav-test-link-2').closest('li');
        const unselected = screen.getByTestId('global_nav-test-link').closest('li');
        expect(selected).toHaveClass('bg-primary');
        expect(unselected).not.toHaveClass('bg-primary');
        expect(screen.getByTestId('global_nav-test-link')).not.toHaveAttribute('aria-current');
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

describe('MainNav collapsed item content', () => {
    const collapsibleNavData: MainNavData = {
        ...mainNavData,
        secondaryList: [
            {
                label: 'External Link',
                icon: <AppIcon.LineChart size={24} />,
                route: 'https://example.com',
                target: '_blank',
                testId: 'global_nav-test-external',
            },
            {
                label: 'Control Item',
                icon: <AppIcon.LineChart size={24} />,
                onClick: vi.fn(),
                control: <span data-testid='test-control'>control</span>,
                testId: 'global_nav-test-control',
            },
        ],
    };
    const renderNav = (isExpanded: boolean) => {
        createMatchMediaController(isExpanded);
        window.localStorage.setItem(NAV_EXPANDED_STORAGE_KEY, JSON.stringify(isExpanded));
        render(<MainNav mainNavData={collapsibleNavData} />);
    };
    afterEach(() => {
        window.localStorage.clear();
        vi.unstubAllGlobals();
    });
    it('shows the control and external link icon only when expanded', () => {
        renderNav(true);
        expect(screen.getByTestId('test-control')).toBeInTheDocument();
        expect(
            within(screen.getByTestId('global_nav-test-external')).getByText('arrow-up-right-from-square')
        ).toBeInTheDocument();
    });
    it('hides the control and external link icon when collapsed', () => {
        renderNav(false);
        expect(screen.queryByTestId('test-control')).not.toBeInTheDocument();
        expect(
            within(screen.getByTestId('global_nav-test-external')).queryByText('arrow-up-right-from-square')
        ).not.toBeInTheDocument();
    });
    it('sizes the logo container to full width when expanded', () => {
        renderNav(true);
        expect(screen.getByTestId('global_nav-logo')).toHaveClass('w-[166px]');
        expect(screen.getByTestId('global_nav-logo')).not.toHaveClass('w-10');
    });
    it('sizes the logo container when collapsed', () => {
        renderNav(false);
        expect(screen.getByTestId('global_nav-logo')).toHaveClass('w-10');
    });
    it('sizes the home link to fit the logo', () => {
        renderNav(false);
        expect(screen.getByRole('link', { name: 'BloodHound home' })).toHaveClass('w-fit');
    });
    it('shows the footer when expanded', () => {
        renderNav(true);
        expect(screen.getByTestId('global_nav-footer')).toHaveClass('visible', 'opacity-100', 'h-[58px]');
    });
    it('hides the footer when collapsed', () => {
        renderNav(false);
        expect(screen.getByTestId('global_nav-footer')).toHaveClass('invisible', 'opacity-0', 'h-0');
    });
});

describe('MainNav label typography', () => {
    it('uses h5 for primary items and body1 for secondary items', () => {
        render(<MainNav mainNavData={mainNavData} />);

        const primaryLabel = within(screen.getByTestId('global_nav-primary-list')).getAllByTestId(
            'global_nav-item-label-text'
        )[0];
        const secondaryLabel = within(screen.getByTestId('global_nav-secondary-list')).getByTestId(
            'global_nav-item-label-text'
        );
        expect(primaryLabel).toHaveClass('font-bold', 'text-lg');
        expect(secondaryLabel).toHaveClass('font-normal', 'text-base');
    });
});
describe('MainNav subnav item', () => {
    const subNavData: MainNavData = {
        ...mainNavData,
        secondaryList: [
            {
                label: 'Administration',
                icon: <AppIcon.LineChart size={24} />,
                testId: 'global_nav-test-subnav',
                subNav: [
                    {
                        title: 'Section',
                        items: [{ label: 'Child', path: '/administration/child' } as SubNavItem],
                    },
                ],
            },
        ],
    };
    it.each(['/administration/child', '/administration/child/details'])(
        'highlights the subnav trigger at %s',
        (pathname) => {
            render(<MainNav mainNavData={subNavData} />, { route: pathname });
            expect(screen.getByTestId('global_nav-test-subnav').closest('li')).toHaveClass('bg-primary');
        }
    );
    it.each(['/test', '/administration/child-other'])('does not highlight the subnav trigger at %s', (pathname) => {
        render(<MainNav mainNavData={subNavData} />, { route: pathname });
        expect(screen.getByTestId('global_nav-test-subnav').closest('li')).not.toHaveClass('bg-primary');
    });

    it('opens the subnav when the trigger is clicked', async () => {
        const user = userEvent.setup();
        render(<MainNav mainNavData={subNavData} />);
        const trigger = screen.getByTestId('global_nav-test-subnav');
        expect(trigger).toHaveAttribute('aria-expanded', 'false');

        await user.click(trigger);

        expect(trigger).toHaveAttribute('aria-expanded', 'true');
        expect(await screen.findByTestId('sub-nav')).toBeInTheDocument();
        expect(screen.getByText('Child')).toBeInTheDocument();
    });
});
