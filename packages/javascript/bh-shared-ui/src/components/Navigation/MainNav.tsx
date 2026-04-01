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
import { MainNavData, MainNavDataListItem, MainNavLogoDataObject } from './types';

const isExpandedStorageKey = 'isNavExpanded';

export const MainNavLogo: FC<{ data: MainNavLogoDataObject }> = ({ data }) => {
    return (
        <div className='pt-2 overflow-hidden' data-testid='global_nav-home'>
            <AppLink to={data.project.route}>{data.project.icon}</AppLink>
        </div>
    );
};

const MainNavListItem: FC<{ isExpanded: boolean; item: MainNavDataListItem }> = ({ isExpanded, item }) => {
    const location = useLocation();
    const [isSubNavOpen, setIsSubNavOpen] = useState(false);
    const navItemRef = useRef<HTMLLIElement>(null);
    const { control, icon, label, route, subNav, target, testId } = item;

    const isActiveRoute = route ? location.pathname.includes(route.replace(/\*/g, '')) : false;
    const isActiveSubNavRoute = subNav
        ? subNav.some((section) => section.items.some((item) => location.pathname.includes(item.path)))
        : false;
    const isSubNavVisible = subNav && isSubNavOpen;

    const navItemContainerClasses = cn('h-10 text-xl px-2 rounded overflow-hidden flex items-center cursor-pointer', {
        'text-primary dark:text-[#8D8BF8] bg-neutral-4': isActiveRoute || isActiveSubNavRoute,
        'group hover:text-primary-variant hover:dark:text-[#7B78FD] hover:bg-neutral-3 dark:hover:bg-neutral-2':
            !isActiveRoute && !isActiveSubNavRoute,
    });

    // Full width ensures that even clicking white space activates menu item
    const navItemClasses = 'w-full flex items-center gap-x-2 group-hover:cursor-pointer';

    const labelElement = (
        <span className='whitespace-nowrap flex items-center gap-x-2'>
            <span data-testid='global_nav-item-label-icon'>{icon}</span>
            <span data-testid='global_nav-item-label-text'>{label}</span>
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
                <li className={navItemContainerClasses} ref={navItemRef}>
                    {navItem}
                </li>
            </ConditionalTooltip>

            {isSubNavVisible &&
                createPortal(
                    <SubNav close={closeSubNav} isExpanded={isExpanded} sections={subNav} triggerRef={navItemRef} />,
                    document.body
                )}
        </>
    );
};

const MainNavFooter: FC<{ mainNavData: MainNavData }> = ({ mainNavData }) => {
    const { data: apiVersionResponse, isSuccess } = useApiVersion();
    const apiVersion = isSuccess && apiVersionResponse?.server_version;
    const logoImage = mainNavData.logo.specterOps.image;

    return (
        <div className='py-3 text-xs overflow-hidden'>
            {/* Container div keeps footer content centered */}
            <div className='flex flex-col w-[264px] items-center gap-2'>
                {/* App version */}
                <div data-testid='global_nav-version-number'>BloodHound: {apiVersion}</div>

                {/* SpecterOps logo */}
                <div className='flex items-center gap-1' data-testid='global_nav-powered-by'>
                    powered by
                    <img
                        src={logoImage.imageUrl}
                        alt={logoImage.altText}
                        height={logoImage.dimensions.height}
                        width={logoImage.dimensions.width}
                        className={logoImage.classes}
                    />
                </div>
            </div>
        </div>
    );
};

const MainNav: FC<{ mainNavData: MainNavData }> = ({ mainNavData }) => {
    const [isExpanded, setIsExpanded] = useLocalStorage<boolean>(isExpandedStorageKey, true);
    const handleToggleNav = () => setIsExpanded(!isExpanded);
    const navigate = useAppNavigate();

    const keybindings = useMemo(
        () =>
            [...mainNavData.primaryList, ...mainNavData.secondaryList]
                .filter((navItem) => !!navItem.route)
                .reduce((acc, curr, index) => {
                    return {
                        ...acc,
                        [`Digit${index + 1}`]: () => navigate(curr.route!),
                    };
                }, {}),
        [mainNavData, navigate]
    );

    useKeybindings(keybindings);

    return (
        <nav
            className={cn(
                'flex flex-col w-nav-width h-full p-2 font-medium shadow-md z-nav print:hidden',
                'transition-all duration-300 ease-in',
                'bg-[#F2F2F2] dark:bg-[#1F1F1F]',
                { 'w-nav-width-expanded': isExpanded }
            )}>
            {/* Bloodhound logo */}
            <MainNavLogo data={mainNavData.logo} />

            {/* Nav expand/collapse button */}
            <div className='text-right z-navToggle'>
                <Button
                    aria-label='Toggle Navigation'
                    // Negative right margin allows button to hover outside nav bar bounds
                    className={cn(
                        'w-6 h-6 -mr-5 border-none',
                        'text-[#121212] dark:text-white',
                        'bg-neutral-4 dark:bg-neutral-5',
                        'hover:bg-[#B2B8BE] dark:hover:bg-neutral-3',
                        'active:bg-[#C0C6CB] dark:active:bg-neutral-2',
                        'focus:text-white focus:bg-secondary dark:focus:text-[#121212] dark:focus:bg-[#6F7DFF]',
                        'focus:ring-2 focus:ring-offset-2 focus:ring-secondary dark:focus:ring-offset-[#1F1F1F] dark:focus:ring-[#6F7DFF]',
                        { 'rotate-180': isExpanded }
                    )}
                    onClick={handleToggleNav}
                    variant='icon'>
                    <FontAwesomeIcon icon={faCaretRight} />
                </Button>
            </div>

            {/* Nav menu top and bottom lists of items */}
            <ul className='flex flex-col gap-4' data-testid='global_nav-primary-list'>
                {mainNavData.primaryList.map((item: MainNavDataListItem, index: number) => (
                    <MainNavListItem item={item} isExpanded={isExpanded} key={index} />
                ))}
            </ul>

            {/* Spacer to push footer to bottom */}
            <div className='flex-1' />

            <ul className='flex flex-col gap-4' data-testid='global_nav-secondary-list'>
                {mainNavData.secondaryList.map((item: MainNavDataListItem, index: number) => (
                    <MainNavListItem item={item} isExpanded={isExpanded} key={index} />
                ))}
            </ul>

            <MainNavFooter mainNavData={mainNavData} />
        </nav>
    );
};

export default MainNav;
