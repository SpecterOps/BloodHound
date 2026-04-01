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

import { FC, useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { useOnClickOutside } from '../../hooks';
import { SubNavItem, SubNavSection } from '../../types';
import { cn } from '../../utils';
import { AppLink } from './AppLink';

const SubNavListItem: FC<{ item: Pick<SubNavItem, 'label' | 'path'> }> = ({ item }) => {
    const location = useLocation();
    const { label, path } = item;
    const isActiveRoute = path ? location.pathname.includes(path.replace(/\*/g, '')) : false;

    return (
        <li
            className={cn('mx-2 px-2 py-0.5 rounded', {
                'text-primary dark:text-[#8D8BF8] bg-neutral-4': isActiveRoute,
                'hover:text-primary-variant hover:dark:text-[#7B78FD] hover:bg-neutral-3 dark:hover:bg-neutral-2':
                    !isActiveRoute,
            })}>
            {/* Full width ensures that even clicking white space activates the link */}
            {/* Anchor uses block display instead of inline so full width works */}
            <AppLink className='w-full block' to={path}>
                {label}
            </AppLink>
        </li>
    );
};

type SubNavSections = Omit<SubNavSection, 'order' | 'items'> & {
    items: Pick<SubNavItem, 'label' | 'path'>[];
};

interface SubNavProps {
    isExpanded: boolean;
    onNavigate: () => void;
    sections: SubNavSections[];
}

const SubNav: React.FC<SubNavProps> = ({ isExpanded, onNavigate, sections }) => {
    // Handles slide-in transition
    const [visible, setVisible] = useState(false);

    useEffect(() => {
        requestAnimationFrame(() => setVisible(true));
    }, []);

    // ref used to close subnav when clicking outside of it
    const ref = useRef<HTMLDivElement>(null);
    useOnClickOutside(ref, onNavigate);

    // subnav also closes when mouse leaves
    const handleMouseExit = () => {
        onNavigate();
    };

    return (
        <nav
            className={cn(
                'bottom-28 py-2 rounded-lg cursor-default z-subNav',
                'flex flex-col gap-8 absolute shadow-md',
                'bg-[#F2F2F2] dark:bg-[#1F1F1F]',
                'transition-all duration-300 ease-out',
                {
                    'opacity-100': visible,
                    'opacity-0': !visible,
                    'left-subnav-expanded': isExpanded,
                    'left-subnav-collapsed': !isExpanded,
                }
            )}
            data-testid='sub-nav'
            ref={ref}
            onMouseLeave={handleMouseExit}>
            {sections.map((section, sectionIndex) => (
                <ul key={sectionIndex} className='flex flex-col gap-1'>
                    {/* Section title */}
                    <li className='px-4 text-lg font-medium'>{section.title}</li>

                    {/* Section items */}
                    {section.items.map((item, itemIndex) => (
                        <SubNavListItem key={itemIndex} item={item} />
                    ))}
                </ul>
            ))}
        </nav>
    );
};

export default SubNav;
