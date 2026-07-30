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
import { useApiVersion, useKeybindings, useNavExpanded } from '../../hooks';
import { privilegeZonesPath } from '../../routes';
import { cn, useAppNavigate } from '../../utils';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';
import { ConditionalTooltip } from '../ConditionalTooltip';
import { AppLink } from './AppLink';
import SubNav from './SubNav';
import type { MainNavData, MainNavDataListItem, MainNavLogoDataObject, NavActionItem, NavLinkItem } from './types';

export const MainNavLogo: FC<{
    data: MainNavLogoDataObject;
    visualRefresh: boolean;
}> = ({ data, visualRefresh }) => {
    return (
        <div
            className={cn({
                'flex-none basis-10 m-2 mt-4 mb-6 overflow-hidden': !visualRefresh,
                'flex-none h-[72px] overflow-hidden border-b border-side-nav-divider px-2 pt-4 text-side-nav-icon-text [&_a]:block [&_a]:h-[30px] [&_a]:w-full [&_a]:overflow-hidden [&_svg]:!h-[30px] [&_svg]:!w-[166px]':
                    visualRefresh,
            })}
            data-testid='global_nav-home'>
            <AppLink to={data.project.route}>{data.project.icon}</AppLink>
        </div>
    );
};

const MainNavListItem: FC<{
    /** Whether the main nav is in its expanded (wide) state; controls tooltip visibility */
    isExpanded: boolean;
    /** The navigation item data to render, either a link, action, or subnav trigger */
    item: MainNavDataListItem;
    /** The visual hierarchy used by the top-level or bottom utility navigation */
    variant: 'primary' | 'utility';
    /** Whether the BHE visual refresh is enabled */
    visualRefresh: boolean;
}> = ({ isExpanded, item, variant, visualRefresh }) => {
    const location = useLocation();
    const [isSubNavOpen, setIsSubNavOpen] = useState(false);
    const navItemRef = useRef<HTMLLIElement>(null);
    const { control, icon, label, route, subNav, target, testId } = item;

    const isActiveRoute = route ? location.pathname.includes(route.replace(/\*/g, '')) : false;
    const isActiveSubNavRoute = subNav
        ? subNav.some((section) => section.items.some((item) => location.pathname.includes(item.path)))
        : false;
    const isSubNavVisible = subNav && isSubNavOpen;

    const navItemContainerClasses = cn('rounded flex items-center cursor-pointer', {
        'group text-side-nav-icon-text hover:bg-side-nav-item-hover active:bg-side-nav-item-active focus-within:shadow-side-nav-focus':
            visualRefresh,
        'text-xl': !visualRefresh,
        'text-primary dark:text-[#8D8BF8] bg-neutral-4': !visualRefresh && (isActiveRoute || isActiveSubNavRoute),
        'group hover:text-primary-variant hover:dark:text-[#7B78FD] hover:bg-neutral-3 dark:hover:bg-[#1A1A1A]':
            !visualRefresh && !isActiveRoute && !isActiveSubNavRoute,
        'bg-side-nav-item-active': visualRefresh && (isActiveRoute || isActiveSubNavRoute),
        'font-heading text-lg font-bold leading-5 tracking-[.25px]': visualRefresh && variant === 'primary',
        'font-sans text-base font-normal leading-6': visualRefresh && variant === 'utility',
    });

    // Full width ensures that even clicking white space activates menu item
    const navItemClasses = cn('w-full px-2 flex items-center gap-x-2 group-hover:cursor-pointer', {
        'h-10': !visualRefresh,
        'h-8 py-1 outline-none': visualRefresh,
    });

    const labelElement = (
        <span className='whitespace-nowrap flex items-center gap-x-2'>
            <span data-testid='global_nav-item-label-icon'>{icon}</span>
            <span
                className={cn({ hidden: visualRefresh && !isExpanded })}
                data-testid='global_nav-item-label-text'
                aria-label={label}>
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
            to={route}
            aria-label={visualRefresh && !isExpanded ? label : undefined}
            aria-current={visualRefresh && isActiveRoute ? 'page' : undefined}>
            {labelElement}
            {target === '_blank' && (
                <FontAwesomeIcon
                    className={cn({ hidden: visualRefresh && !isExpanded })}
                    icon={faExternalLink}
                    size='sm'
                />
            )}
        </AppLink>
    ) : (
        <div
            className={navItemClasses}
            data-testid={testId}
            onClick={onClick}
            onKeyDown={onKeyDown}
            role='button'
            aria-label={visualRefresh && !isExpanded ? label : undefined}
            aria-expanded={visualRefresh && subNav ? isSubNavOpen : undefined}
            tabIndex={0}>
            {labelElement}
            {control && <span className={cn('ml-1', { hidden: visualRefresh && !isExpanded })}>{control}</span>}
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
                    <SubNav
                        close={closeSubNav}
                        isExpanded={isExpanded}
                        sections={subNav}
                        triggerRef={navItemRef}
                        visualRefresh={visualRefresh}
                    />,
                    navItemRef.current.closest('nav')!
                )}
        </>
    );
};

const MainNavFooter: FC<{
    /** Object containing image props */
    image: MainNavLogoDataObject['specterOps']['image'];
    /** Whether the main nav is expanded */
    isExpanded: boolean;
    /** Whether the BHE visual refresh is enabled */
    visualRefresh: boolean;
}> = ({ image, isExpanded, visualRefresh }) => {
    const { data: apiVersionResponse, isSuccess } = useApiVersion();
    const apiVersion = isSuccess && apiVersionResponse?.server_version;

    if (visualRefresh && !isExpanded) {
        return null;
    }

    return (
        <div
            className={cn({
                'py-3 text-xs': !visualRefresh,
                'w-[264px] border-t border-side-nav-divider py-2 pl-4 pr-2 font-sans text-side-nav-icon-text':
                    visualRefresh,
            })}>
            <div
                className={cn('flex flex-col', {
                    'w-[264px] items-center gap-2': !visualRefresh,
                    'items-start': visualRefresh,
                })}>
                {/* App version */}
                <div
                    className={cn({
                        'text-[13px] font-medium leading-[22px] tracking-[.25px] text-side-nav-version': visualRefresh,
                    })}
                    data-testid='global_nav-version-number'>
                    BloodHound: {apiVersion}
                </div>

                {/* SpecterOps logo */}
                <div
                    className={cn('flex items-center gap-1', {
                        'text-xs leading-5 tracking-[.25px]': visualRefresh,
                    })}
                    data-testid='global_nav-powered-by'>
                    {visualRefresh ? 'Powered by' : 'powered by'}
                    <img
                        src={image.imageUrl}
                        alt={image.altText}
                        height={image.dimensions.height}
                        width={image.dimensions.width}
                        className={cn({ 'h-4 w-[124px] object-contain': visualRefresh }, image.classes)}
                    />
                </div>
            </div>
        </div>
    );
};

const MainNav: FC<{ mainNavData: MainNavData; visualRefresh?: boolean }> = ({ mainNavData, visualRefresh = false }) => {
    const [isExpanded, setIsExpanded] = useNavExpanded();
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
                className={cn('absolute border-none z-navToggle', 'transition-all duration-300 ease-in', {
                    'top-14 w-6 h-6 text-main bg-neutral-4 dark:bg-neutral-5 hover:bg-[#B2B8BE] dark:hover:bg-neutral-3 active:ring-0 active:bg-[#C0C6CB] dark:active:bg-neutral-2 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-focus focus-visible:ring-offset-focus-offset':
                        !visualRefresh,
                    'top-14 size-8 text-side-nav-icon-text bg-side-nav-toggle-default hover:bg-side-nav-toggle-hover active:ring-0 active:bg-side-nav-toggle-active focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:shadow-side-nav-focus':
                        visualRefresh,
                    'rotate-180 left-[16.75rem]': isExpanded && !visualRefresh,
                    'rotate-180 left-[15.5rem]': isExpanded && visualRefresh,
                    'left-[2.75rem]': !isExpanded && !visualRefresh,
                    'left-10': !isExpanded && visualRefresh,
                })}
                onClick={handleToggleNav}
                variant='icon'>
                <FontAwesomeIcon icon={faCaretRight} />
            </Button>

            <nav
                className={cn(
                    'flex flex-col flex-none shadow-md z-nav print:hidden overflow-hidden',
                    'transition-all duration-300 ease-in',
                    {
                        'font-medium bg-[#F2F2F2] dark:bg-[#1F1F1F]': !visualRefresh,
                        'bg-side-nav-bg text-side-nav-icon-text dark:bg-[#1F1F1F]': visualRefresh,
                        'basis-nav-width': !isExpanded,
                        'basis-[17.5rem]': isExpanded && !visualRefresh,
                        'basis-nav-width-expanded': isExpanded && visualRefresh,
                    }
                )}>
                {/* Bloodhound logo */}
                <MainNavLogo data={mainNavData.logo} visualRefresh={visualRefresh} />

                <div
                    className={cn('flex flex-col h-full overflow-x-hidden overflow-y-auto', {
                        'mx-2': !visualRefresh,
                    })}>
                    {/* Nav menu top and bottom lists of items */}
                    <ul
                        className={cn('flex flex-col flex-grow', {
                            'gap-2': !visualRefresh,
                            'gap-4 px-2 py-4': visualRefresh,
                        })}
                        data-testid='global_nav-primary-list'>
                        {mainNavData.primaryList.map((item: MainNavDataListItem) => (
                            <MainNavListItem
                                item={item}
                                isExpanded={isExpanded}
                                variant='primary'
                                visualRefresh={visualRefresh}
                                key={item.testId}
                            />
                        ))}
                    </ul>

                    <ul
                        className={cn('flex flex-col gap-2', {
                            'mt-2': !visualRefresh,
                            'px-2 py-4': visualRefresh,
                        })}
                        data-testid='global_nav-secondary-list'>
                        {mainNavData.secondaryList.map((item: MainNavDataListItem) => (
                            <MainNavListItem
                                item={item}
                                isExpanded={isExpanded}
                                variant='utility'
                                visualRefresh={visualRefresh}
                                key={item.testId}
                            />
                        ))}
                    </ul>

                    <MainNavFooter
                        image={mainNavData.logo.specterOps.image}
                        isExpanded={isExpanded}
                        visualRefresh={visualRefresh}
                    />
                </div>
            </nav>
        </>
    );
};

export default MainNav;
