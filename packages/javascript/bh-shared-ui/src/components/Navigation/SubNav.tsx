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

import { Typography, TypographyVariants } from 'doodle-ui';
import { FC, RefObject, useCallback, useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { useOnClickOutside } from '../../hooks';
import { SubNavItem, SubNavSection } from '../../types';
import { cn } from '../../utils';
import { AppLink } from './AppLink';
import { isRouteActive } from './utils';

const SubNavListItem: FC<{ item: Pick<SubNavItem, 'label' | 'path'> }> = ({ item }) => {
    const location = useLocation();
    const { label, path } = item;
    const isActiveRoute = isRouteActive(location.pathname, path);

    return (
        <li
            className={cn(
                'min-h-7 rounded focus-within:focus-ring-inset focus-within:[--focus-ring:var(--common-white)]',
                {
                    'text-common-white bg-primary dark:text-primary-variant dark:bg-neutral-4': isActiveRoute,
                    'hover:text-common-white hover:bg-secondary dark:hover:text-[#7B78FD] dark:hover:bg-neutral-200':
                        !isActiveRoute,
                }
            )}>
            {/* Full width ensures that even clicking white space activates the link */}
            {/* Anchor uses block display instead of inline so full width works */}
            <AppLink
                aria-current={isActiveRoute ? 'page' : undefined}
                className={cn(
                    'w-full block px-2 py-0.5 outline-none',
                    TypographyVariants({ variant: 'body2' }),
                    'truncate !text-inherit'
                )}
                to={path}>
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
}> = ({ isExpanded, close, sections, triggerRef }) => {
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
                'bottom-2 w-[264px] p-2 rounded cursor-default z-subNav',
                'flex flex-col gap-8 absolute',
                'text-common-white bg-primary-variant dark:text-main dark:bg-neutral-300',
                'transform-gpu translate-z-[0px]', // This line addresses a Safari hardware rendering bug
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
            onMouseLeave={close}>
            {sections.map((section, sectionIndex) => (
                <ul key={sectionIndex} className='flex flex-col gap-2'>
                    {/* Section title */}
                    <Typography className='px-2 !text-inherit' component='li' variant='h6'>
                        {section.title}
                    </Typography>
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
