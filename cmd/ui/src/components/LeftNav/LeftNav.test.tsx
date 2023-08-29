// Copyright 2023 Specter Ops, Inc.
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

import { render, screen } from 'src/test-utils';
import LeftNav from 'src/components/LeftNav';

describe('LeftNav', () => {
    it('should render a single section with a single navigation item', () => {
        const testSectionTitle = 'testSectionTitle';
        const testNavItem = {
            label: 'testNavItemLabel',
            path: '/administration/test-nav-item-path',
        };

        render(<LeftNav sections={[{ title: testSectionTitle, items: [testNavItem] }]} />);
        expect(screen.getByRole('navigation')).toBeInTheDocument();
        expect(screen.getByText(testSectionTitle)).toBeInTheDocument();
        expect(screen.getByText(testNavItem.label)).toBeInTheDocument();
        expect(screen.getByRole('link', { name: testNavItem.label })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: testNavItem.label })).toHaveAttribute('href', testNavItem.path);
    });

    it('should render many sections with a many navigation items', () => {
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

        render(<LeftNav sections={testSections} />);
        expect(screen.getByRole('navigation')).toBeInTheDocument();
        expect(screen.getAllByRole('link')).toHaveLength(9);
        for (const section of testSections) {
            expect(screen.getByText(section.title)).toBeInTheDocument();

            for (const item of section.items) {
                expect(screen.getByText(item.label)).toBeInTheDocument();
                expect(screen.getByRole('link', { name: item.label })).toBeInTheDocument();
                expect(screen.getByRole('link', { name: item.label })).toHaveAttribute('href', item.path);
            }
        }
    });
});
