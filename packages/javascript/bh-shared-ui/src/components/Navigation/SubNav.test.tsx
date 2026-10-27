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
import { render, screen } from '../../test-utils';
import SubNav from './SubNav';

describe('SubNav', () => {
    it('should render a single section with a single navigation item', () => {
        const testSectionTitle = 'testSectionTitle';
        const testNavItem = {
            label: 'testNavItemLabel',
            path: '/administration/test-nav-item-path',
        };

        render(
            <SubNav
                close={() => {}}
                isExpanded={false}
                sections={[{ title: testSectionTitle, items: [testNavItem] }]}
            />
        );
        expect(screen.getByRole('navigation')).toBeInTheDocument();
        expect(screen.getByText(testSectionTitle)).toBeInTheDocument();
        expect(screen.getByText(testNavItem.label)).toBeInTheDocument();
        expect(
            screen.getByRole('link', { name: `Navigate to ${testNavItem.path.replace(/\//g, ' ').trim()}` })
        ).toBeInTheDocument();
        expect(
            screen.getByRole('link', { name: `Navigate to ${testNavItem.path.replace(/\//g, ' ').trim()}` })
        ).toHaveAttribute('href', testNavItem.path);
    });

    it('should render many sections with many navigation items', () => {
        const testSections = [
            {
                title: 'testSection1',
                items: [
                    { label: 'testNavItem1-1', path: '/administration/test-nav-item-1-1' },
                    { label: 'testNavItem1-2', path: '/administration/test-nav-item-1-2' },
                    { label: 'testNavItem1-3', path: '/administration/test-nav-item-1-3' },
                ],
            },
            {
                title: 'testSection2',
                items: [
                    { label: 'testNavItem2-1', path: '/administration/test-nav-item-2-1' },
                    { label: 'testNavItem2-2', path: '/administration/test-nav-item-2-2' },
                    { label: 'testNavItem2-3', path: '/administration/test-nav-item-2-3' },
                ],
            },
            {
                title: 'testSection3',
                items: [
                    { label: 'testNavItem3-1', path: '/administration/test-nav-item-3-1' },
                    { label: 'testNavItem3-2', path: '/administration/test-nav-item-3-2' },
                    { label: 'testNavItem3-3', path: '/administration/test-nav-item-3-3' },
                ],
            },
        ];

        render(<SubNav close={() => {}} isExpanded={false} sections={testSections} />);
        expect(screen.getByRole('navigation')).toBeInTheDocument();
        expect(screen.getAllByRole('link')).toHaveLength(9);
        for (const section of testSections) {
            expect(screen.getByText(section.title)).toBeInTheDocument();

            for (const item of section.items) {
                expect(screen.getByText(item.label)).toBeInTheDocument();
                expect(
                    screen.getByRole('link', { name: `Navigate to ${item.path.replace(/\//g, ' ').trim()}` })
                ).toBeInTheDocument();
                expect(
                    screen.getByRole('link', { name: `Navigate to ${item.path.replace(/\//g, ' ').trim()}` })
                ).toHaveAttribute('href', item.path);
            }
        }
    });
});

describe('SubNav styling', () => {
    const sections = [
        {
            title: 'Section Title',
            items: [
                { label: 'Active', path: '/administration/active' },
                { label: 'Inactive', path: '/administration/inactive' },
            ],
        },
    ];
    it('highlights only the active route item', () => {
        render(<SubNav close={() => {}} isExpanded={false} sections={sections} />, {
            route: '/administration/active',
        });
        expect(screen.getByText('Active').closest('li')).toHaveClass('bg-primary');
        expect(screen.getByText('Inactive').closest('li')).not.toHaveClass('bg-primary');
        expect(screen.getByText('Active')).toHaveAttribute('aria-current', 'page');
        expect(screen.getByText('Inactive')).not.toHaveAttribute('aria-current');
    });
    it('highlights the item on a child route', () => {
        render(<SubNav close={() => {}} isExpanded={false} sections={sections} />, {
            route: '/administration/active/child',
        });
        expect(screen.getByText('Active').closest('li')).toHaveClass('bg-primary');
        expect(screen.getByText('Active')).toHaveAttribute('aria-current', 'page');
    });
    it('does not highlight an item that only shares a prefix with the current path', () => {
        render(<SubNav close={() => {}} isExpanded={false} sections={sections} />, {
            route: '/administration/active-other',
        });
        expect(screen.getByText('Active').closest('li')).not.toHaveClass('bg-primary');
        expect(screen.getByText('Active')).not.toHaveAttribute('aria-current');
    });
    it('renders the section title as an h6 list item', () => {
        render(<SubNav close={() => {}} isExpanded={false} sections={sections} />);
        const title = screen.getByText('Section Title');
        expect(title.tagName).toBe('LI');
        expect(title).toHaveClass('font-semibold');
    });
    it.each([
        [true, 'left-subnav-expanded'],
        [false, 'left-subnav-collapsed'],
    ])('positions based on isExpanded=%s', (isExpanded, expectedClass) => {
        render(<SubNav close={() => {}} isExpanded={isExpanded} sections={sections} />);
        expect(screen.getByTestId('sub-nav')).toHaveClass(expectedClass);
    });
    it('calls close on mouse leave', async () => {
        const close = vi.fn();
        const user = userEvent.setup();
        render(<SubNav close={close} isExpanded={false} sections={sections} />);
        await user.hover(screen.getByTestId('sub-nav'));
        await user.unhover(screen.getByTestId('sub-nav'));
        expect(close).toHaveBeenCalled();
    });
});
