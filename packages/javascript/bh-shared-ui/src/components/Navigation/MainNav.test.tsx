// Copyright 2025 Specter Ops, Inc.
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
import { BrowserRouter } from 'react-router-dom';
import { render, screen, within } from '../../test-utils';
import { AppIcon } from '../AppIcon';
import MainNav from './MainNav';
import { MainNavData, MainNavDataListItem, MainNavLogoDataObject } from './types';

const MainNavLogoData: MainNavLogoDataObject = {
    project: {
        route: '/',
        icon: <AppIcon.BHCELogo size={24} />,
        image: {
            imageUrl: `/test`,
            dimensions: { height: '40px', width: '165px' },
            classes: 'ml-4',
            altText: 'BHE Text Logo',
        },
    },
    specterOps: {
        image: {
            imageUrl: `/test`,
            dimensions: { height: '40px', width: '165px' },
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
];

const handleClick = vi.fn();

const MainNavSecondaryListData: MainNavDataListItem[] = [
    {
        label: 'Action Item',
        icon: <AppIcon.LineChart size={24} />,
        functionHandler: handleClick,
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
        render(
            <BrowserRouter>
                <MainNav mainNavData={mainNavData} />
            </BrowserRouter>
        );
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
        const linkItem = await within(primaryList).findByRole('link');
        const linkItemIcon = await within(primaryList).findByTestId('global_nav-item-label-icon');
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

        expect(testLinkItem.functionHandler).toBeCalled();
    });
    it('should render a label and version number when expanded', async () => {
        const MainNavBar = await screen.findByRole('navigation');
        expect(MainNavBar).toHaveClass('group');

        const versionNumberContainer = await within(MainNavBar).findByTestId('main-nav-version-number');
        const versionNumberLabel = await within(versionNumberContainer).findByText(
            `BloodHound: ${currentVersionNumber}`
        );

        // ---- collapsed classes ----
        expect(versionNumberLabel).toHaveClass('hidden');
        expect(versionNumberLabel).toHaveClass('opacity-0');
        // ---- collapsed classes ----

        // ---- classes displayed on hover ----
        expect(versionNumberLabel).toHaveClass('group-hover:opacity-100');
        expect(versionNumberLabel).toHaveClass('group-hover:block');
        // ---- classes displayed on hover ----
    });
    it('should only render an icon in list item when collapsed and the label should be styled to be hidden but appear on group-hover of the nav', async () => {
        const testLinkItem = MainNavPrimaryListData[0];

        const MainNavBar = screen.getByRole('navigation');
        expect(MainNavBar).toHaveClass('group');

        const primaryList = await within(MainNavBar).findByTestId('global_nav-primary-list');
        const linkItemIcon = await within(primaryList).findByTestId('global_nav-item-label-icon');
        const linkItemText = await within(primaryList).findByText(testLinkItem.label as string);

        expect(linkItemIcon).toBeInTheDocument();

        // ---- collapsed classes ----
        expect(linkItemText).toHaveClass('hidden');
        expect(linkItemText).toHaveClass('opacity-0');
        // ---- collapsed classes ----

        // ---- classes displayed on hover ----
        expect(linkItemText).toHaveClass('group-hover:opacity-100');
        expect(linkItemText).toHaveClass('group-hover:flex');
        // ---- classes displayed on hover ----
    });
    it('should style the powered-by to display when nav is expanded', async () => {
        const MainNavBar = screen.getByRole('navigation');
        expect(MainNavBar).toHaveClass('group');

        const poweredByTextContainer = await within(MainNavBar).findByTestId('global_nav-powered-by');
        const poweredByText = await within(poweredByTextContainer).findByText(/powered by/i);
        expect(poweredByText).toBeInTheDocument();

        // ---- collapsed classes ----
        expect(poweredByText).toHaveClass('hidden');
        expect(poweredByText).toHaveClass('opacity-0');
        // ---- collapsed classes ----

        // ---- classes displayed on hover ----
        expect(poweredByText).toHaveClass('group-hover:opacity-100');
        expect(poweredByText).toHaveClass('group-hover:flex');
        // ---- classes displayed on hover ----
    });
});
