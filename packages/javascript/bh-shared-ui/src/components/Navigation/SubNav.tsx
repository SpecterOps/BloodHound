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

import { FC, RefObject, useCallback, useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { useOnClickOutside } from '../../hooks';
import { SubNavItem, SubNavSection } from '../../types';
import { cn } from '../../utils';
import { AppLink } from './AppLink';

const SubNavListItem: FC<{ item: Pick<SubNavItem, 'label' | 'path'>; visualRefresh: boolean }> = ({
    item,
    visualRefresh,
}) => {
    const location = useLocation();
    const { label, path } = item;
    const isActiveRoute = path ? location.pathname.includes(path.replace(/\*/g, '')) : false;

    return (
        <li
            className={cn('rounded', {
                'mx-2': !visualRefresh,
                'text-primary dark:text-[#8D8BF8] bg-neutral-4': !visualRefresh && isActiveRoute,
                'hover:text-primary-variant hover:dark:text-[#7B78FD] hover:bg-neutral-3 dark:hover:bg-[#1A1A1A]':
                    !visualRefresh && !isActiveRoute,
                'min-h-7 font-sans text-sm font-normal leading-[1.375rem] text-side-nav-icon-text hover:bg-side-nav-item-hover active:bg-side-nav-item-active focus-within:shadow-side-nav-focus':
                    visualRefresh,
                'bg-side-nav-item-active': visualRefresh && isActiveRoute,
            })}>
            {/* Full width ensures that even clicking white space activates the link */}
            {/* Anchor uses block display instead of inline so full width works */}
            <AppLink
                className={cn('w-full px-2', {
                    'block py-0.5': !visualRefresh,
                    'flex min-h-7 items-center outline-none': visualRefresh,
                })}
                to={path}
                aria-current={visualRefresh && isActiveRoute ? 'page' : undefined}>
                {label}
            </AppLink>
        </li>
    );
};

type SubNavSections = Omit<SubNavSection, 'order' | 'items'> & {
    items: Pick<SubNavItem, 'label' | 'path'>[];
};

const SubNav: React.FC<{
    /** Whether the main nav is in its expanded (wide) state; used to offset subnav position accordingly */
    isExpanded: boolean;
    /** Callback to close the subnav */
    close: () => void;
    /** The grouped sections of navigation items to render inside the subnav */
    sections: SubNavSections[];
    /** Clicking outside of subnav closes it unless trigger element was clicked; Prevents unintended reopens */
    triggerRef?: RefObject<HTMLElement>;
    /** Whether the BHE visual refresh is enabled */
    visualRefresh?: boolean;
}> = ({ isExpanded, close, sections, triggerRef, visualRefresh = false }) => {
    // Handles slide-in transition
    const [visible, setVisible] = useState(false);

    useEffect(() => {
        requestAnimationFrame(() => setVisible(true));
    }, []);

    const ref = useRef<HTMLDivElement>(null);
    const handleClickOutside = useCallback(
        (e: Event) => {
            // trigger element excluded to prevent unintended reopens
            if (triggerRef?.current?.contains(e.target as Node)) return;
            close();
        },
        [triggerRef, close]
    );
    useOnClickOutside(ref, handleClickOutside);

    return (
        <nav
            className={cn(
                'bottom-2 rounded-lg cursor-default z-subNav',
                'flex flex-col gap-8 absolute shadow-md',
                'transform-gpu translate-z-[0px]', // This line addresses a Safari hardware rendering bug
                'transition-all duration-300 ease-out',
                {
                    'py-2 bg-[#F2F2F2] dark:bg-[#1F1F1F]': !visualRefresh,
                    'w-[264px] p-2 bg-side-nav-bg text-side-nav-icon-text dark:bg-[#1F1F1F]': visualRefresh,
                    'opacity-100': visible,
                    'opacity-0': !visible,
                    'left-subnav-expanded': isExpanded && visualRefresh,
                    'left-[18rem]': isExpanded && !visualRefresh,
                    'left-subnav-collapsed': !isExpanded,
                }
            )}
            data-testid='sub-nav'
            ref={ref}
            onMouseLeave={close}>
            {sections.map((section, sectionIndex) => (
                <ul key={sectionIndex} className={cn('flex flex-col', visualRefresh ? 'gap-2' : 'gap-1')}>
                    {/* Section title */}
                    <li
                        className={cn({
                            'px-4 text-lg font-medium': !visualRefresh,
                            'px-2 font-heading text-base font-semibold leading-[1.125rem] tracking-[.25px]':
                                visualRefresh,
                        })}>
                        {section.title}
                    </li>

                    {/* Section items */}
                    {section.items.map((item, itemIndex) => (
                        <SubNavListItem key={itemIndex} item={item} visualRefresh={visualRefresh} />
                    ))}
                </ul>
            ))}
        </nav>
    );
};

export default SubNav;
