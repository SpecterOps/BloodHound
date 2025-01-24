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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { BrowserRouter } from 'react-router-dom';
import { render, screen, within } from '../../test-utils';
import { AppIcon } from '../AppIcon';
import MainNav from './MainNav';

// To do: Type this
// To do: add test id to nav logo icon and image jsx to test if they show up
const MainNavLogoData: any = {
    route: '/',
    icon: <AppIcon.BHCELogo size={24} />,
    image: {
        imageUrl: `/test`,
        dimensions: { height: '40px', width: '165px' },
        classes: 'ml-4',
        altText: 'BHE Text Logo',
    },
};
const MainNavPrimaryListData: any[] = [
    {
        label: 'Link Item',
        icon: <AppIcon.LineChart size={24} />,
        route: '/test',
    },
];

const handleClick = vi.fn();

const MainNavSecondaryListData: any[] = [
    {
        label: 'Action Item',
        icon: <AppIcon.LineChart size={24} />,
        functionHandler: handleClick,
    },
];

const mainNavData = {
    logo: MainNavLogoData,
    primaryList: MainNavPrimaryListData,
    secondaryList: MainNavSecondaryListData,
};
const server = setupServer(
    rest.get(`/api/version`, async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    data: {
                        API: {
                            current_version: 'v2',
                            deprecated_version: 'v1',
                        },
                        server_version: 'v999.999.999',
                    },
                },
            })
        );
    })
);
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('MainNav', () => {
    beforeEach(() => {
        render(
            <BrowserRouter>
                <MainNav mainNavData={mainNavData} />
            </BrowserRouter>
        );
    });
    it('should render a nav element with logo, two lists and a version number', () => {
        expect(screen.getByRole('navigation')).toBeInTheDocument();
        expect(screen.getByTestId('main-nav-logo')).toBeInTheDocument();
        expect(screen.getByTestId('main-nav-primary-list')).toBeInTheDocument();
        expect(screen.getByTestId('main-nav-secondary-list')).toBeInTheDocument();
        expect(screen.getByTestId('main-nav-version-number')).toBeInTheDocument();
    });
    it('should render a navigation list item', async () => {
        const testLinkItem = MainNavPrimaryListData[0];

        const primaryList = await screen.findByTestId('main-nav-primary-list');
        const linkItem = await within(primaryList).findByRole('link');
        const linkItemIcon = await within(primaryList).findByTestId('main-nav-item-label-icon');
        const linkItemText = await within(primaryList).findByText(testLinkItem.label);

        expect(linkItem).toBeInTheDocument();
        expect(linkItem).toHaveAttribute('href', testLinkItem.route);
        expect(linkItemIcon).toBeInTheDocument();
        expect(linkItemText).toBeInTheDocument();
    });
    it.todo('should render action list item that handles a function', () => {});
    it.todo('should render a version number when collapsed', () => {});
    it.todo('should render a label and version number when expanded', () => {});
    it.todo('should only render an icon in list item when collapsed', () => {});
    it.todo('should render an icon and label in list item when expanded', () => {});
});
