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

import { faCaretRight, faExternalLink } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button } from 'doodle-ui';
import { FC, useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { useLocation } from 'react-router-dom';
import { useApiVersion, useKeybindings, useLocalStorage } from '../../hooks';
import { privilegeZonesPath } from '../../routes';
import { cn, useAppNavigate } from '../../utils';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';
import { ConditionalTooltip } from '../ConditionalTooltip';
import { AppLink } from './AppLink';
import SubNav from './SubNav';
import type { MainNavData, MainNavDataListItem, MainNavLogoDataObject, NavActionItem, NavLinkItem } from './types';

const isExpandedStorageKey = 'isNavExpanded';

export const MainNavLogo: FC<{ data: MainNavLogoDataObject }> = ({ data }) => {
    return (
        <div className='flex-none basis-10 m-2 mt-4 mb-6 overflow-hidden' data-testid='global_nav-home'>
            <AppLink to={data.project.route}>{data.project.icon}</AppLink>
        </div>
    );
};

const MainNavListItem: FC<{
    /** Whether the main nav is in its expanded (wide) state; controls tooltip visibility */
    isExpanded: boolean;
    /** The navigation item data to render, either a link, action, or subnav trigger */
    item: MainNavDataListItem;
}> = ({ isExpanded, item }) => {
    const location = useLocation();
    const [isSubNavOpen, setIsSubNavOpen] = useState(false);
    const navItemRef = useRef<HTMLLIElement>(null);
    const { control, icon, label, route, subNav, target, testId } = item;

    const isActiveRoute = route ? location.pathname.includes(route.replace(/\*/g, '')) : false;
    const isActiveSubNavRoute = subNav
        ? subNav.some((section) => section.items.some((item) => location.pathname.includes(item.path)))
        : false;
    const isSubNavVisible = subNav && isSubNavOpen;

    // Handles nav item text color and background with hover/active interactions
    const navItemContainerClasses = cn('text-xl rounded flex items-center cursor-pointer', {
        'text-primary dark:text-[#8D8BF8] bg-neutral-4': isActiveRoute || isActiveSubNavRoute,
        'group hover:text-primary-variant hover:dark:text-[#7B78FD] hover:bg-neutral-3 dark:hover:bg-[#1A1A1A]':
            !isActiveRoute && !isActiveSubNavRoute,
    });

    // Full width ensures that even clicking white space activates menu item
    const navItemClasses = 'h-10 w-full px-2 flex items-center gap-x-2 group-hover:cursor-pointer';

    const labelElement = (
        <span className='whitespace-nowrap flex items-center gap-x-2'>
            <span data-testid='global_nav-item-label-icon'>{icon}</span>
            <span data-testid='global_nav-item-label-text' aria-label={label}>
                {label}
            </span>
        </span>
    );

    const handleClickSubNav = () => {
        setIsSubNavOpen(!isSubNavOpen);
    };

    const closeSubNav = () => setIsSubNavOpen(false);

    const onClick = subNav ? handleClickSubNav : item.onClick;
    const onKeyDown = adaptClickHandlerToKeyDown(subNav ? handleClickSubNav : item.onClick);

    // If route is defined, render a link otherwise item is an action or subnav item
    const navItem = route ? (
        <AppLink
            className={navItemClasses}
            data-testid={testId}
            // PZ pages discard environment query params so all Zone Objects are counted
            // Some Objects do not have an environmentId (domain sid or tenant id)
            // As such, even using the "all" environments param does not capture everything
            discardQueryParams={route.includes(privilegeZonesPath)}
            target={target}
            to={route}>
            {labelElement}
            {target === '_blank' && <FontAwesomeIcon icon={faExternalLink} size='sm' />}
        </AppLink>
    ) : (
        <div
            className={navItemClasses}
            data-testid={testId}
            onClick={onClick}
            onKeyDown={onKeyDown}
            role='button'
            tabIndex={0}>
            {labelElement}
            {control && <span className='ml-1'>{control}</span>}
        </div>
    );

    return (
        <>
            <ConditionalTooltip condition={!(isExpanded || isSubNavOpen)} side='right' tooltip={label}>
                <li className={navItemContainerClasses} ref={subNav && navItemRef}>
                    {navItem}
                </li>
            </ConditionalTooltip>

            {isSubNavVisible &&
                navItemRef.current?.closest('nav') &&
                createPortal(
                    <SubNav close={closeSubNav} isExpanded={isExpanded} sections={subNav} triggerRef={navItemRef} />,
                    navItemRef.current.closest('nav')!
                )}
        </>
    );
};

const MainNavFooter: FC<{
    /** Object containing image props */
    image: MainNavLogoDataObject['specterOps']['image'];
}> = ({ image }) => {
    const { data: apiVersionResponse, isSuccess } = useApiVersion();
    const apiVersion = isSuccess && apiVersionResponse?.server_version;

    return (
        <div className='py-3 text-xs'>
            {/* Container div keeps footer content centered */}
            <div className='flex flex-col w-[264px] items-center gap-2'>
                {/* App version */}
                <div data-testid='global_nav-version-number'>BloodHound: {apiVersion}</div>

                {/* SpecterOps logo */}
                <div className='flex items-center gap-1' data-testid='global_nav-powered-by'>
                    powered by
                    <img
                        src={image.imageUrl}
                        alt={image.altText}
                        height={image.dimensions.height}
                        width={image.dimensions.width}
                        className={image.classes}
                    />
                </div>
            </div>
        </div>
    );
};

const MainNav: FC<{ mainNavData: MainNavData }> = ({ mainNavData }) => {
    const [isExpanded, setIsExpanded] = useLocalStorage<boolean>(isExpandedStorageKey, true);
    const navigate = useAppNavigate();

    const keybindings = useMemo(
        () =>
            [...mainNavData.primaryList, ...mainNavData.secondaryList]
                .filter((navItem): navItem is NavLinkItem | NavActionItem => !!navItem.route || !!navItem.onClick)
                .reduce((acc, curr, index) => {
                    return {
                        ...acc,
                        [`Digit${index + 1}`]:
                            'route' in curr
                                ? curr.target === '_blank'
                                    ? () => window.open(curr.route, '_blank')
                                    : () => navigate(curr.route)
                                : () => curr.onClick?.(),
                    };
                }, {}),
        [mainNavData, navigate]
    );

    useKeybindings(keybindings);

    const handleToggleNav = () => setIsExpanded(!isExpanded);

    return (
        <>
            {/* Nav expand/collapse button */}
            <Button
                aria-expanded={isExpanded}
                aria-label='Toggle Navigation'
                // Negative right margin allows button to hover outside nav bar bounds
                className={cn(
                    'absolute top-14 w-6 h-6 border-none z-navToggle',
                    'transition-all duration-300 ease-in',
                    'text-[#121212] dark:text-white',
                    'bg-neutral-4 dark:bg-neutral-5',
                    'hover:bg-[#B2B8BE] dark:hover:bg-neutral-3',
                    'active:ring-0 active:bg-[#C0C6CB] dark:active:bg-neutral-2',
                    'focus:text-[#121212] dark:focus:text-white',
                    'focus:ring-2 focus:ring-offset-2 focus:ring-secondary dark:focus:ring-offset-[#1F1F1F]',
                    {
                        'rotate-180 left-[16.75rem]': isExpanded,
                        'left-[2.75rem]': !isExpanded,
                    }
                )}
                onClick={handleToggleNav}
                variant='icon'>
                <FontAwesomeIcon icon={faCaretRight} />
            </Button>

            <nav
                className={cn(
                    'flex flex-col flex-none font-medium shadow-md z-nav print:hidden overflow-hidden',
                    'transition-all duration-300 ease-in',
                    'bg-[#F2F2F2] dark:bg-[#1F1F1F]',
                    { 'basis-nav-width': !isExpanded, 'basis-nav-width-expanded': isExpanded }
                )}>
                {/* Bloodhound logo */}
                <MainNavLogo data={mainNavData.logo} />

                <div className='flex flex-col h-full mx-2 overflow-x-hidden overflow-y-auto'>
                    {/* Nav menu top and bottom lists of items */}
                    <ul className='flex flex-col flex-grow gap-2' data-testid='global_nav-primary-list'>
                        {mainNavData.primaryList.map((item: MainNavDataListItem) => (
                            <MainNavListItem item={item} isExpanded={isExpanded} key={item.testId} />
                        ))}
                    </ul>

                    <ul className='flex flex-col gap-2 mt-2' data-testid='global_nav-secondary-list'>
                        {mainNavData.secondaryList.map((item: MainNavDataListItem) => (
                            <MainNavListItem item={item} isExpanded={isExpanded} key={item.testId} />
                        ))}
                    </ul>

                    <MainNavFooter image={mainNavData.logo.specterOps.image} />
                </div>
            </nav>
        </>
    );
};

export default MainNav;
