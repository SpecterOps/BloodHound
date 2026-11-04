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
import { IconButton, Typography } from 'doodle-ui';
import { FC, useEffect, useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { useLocation } from 'react-router-dom';
import { BREAKPOINTS, useApiVersion, useKeybindings, useMediaQuery, useNavExpanded } from '../../hooks';
import { privilegeZonesPath } from '../../routes';
import { cn, useAppNavigate } from '../../utils';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';
import { ConditionalTooltip } from '../ConditionalTooltip';
import { AppLink } from './AppLink';
import { SkipLink } from './SkipLink';
import SubNav from './SubNav';
import type { MainNavData, MainNavDataListItem, MainNavLogoDataObject, NavActionItem, NavLinkItem } from './types';
import { isRouteActive } from './utils';

export const MainNavLogo: FC<{ data: MainNavLogoDataObject; isExpanded?: boolean }> = ({ data, isExpanded = true }) => {
    return (
        <div className='h-[72px] flex-none border-b border-primary dark:border-neutral-4'>
            <div className='mt-2 h-[46px] w-full overflow-hidden' data-testid='global_nav-home'>
                <AppLink
                    aria-label={'BloodHound home'}
                    className={cn(
                        'ml-1 mt-1 block w-fit overflow-hidden rounded p-1',
                        'focus-visible:outline-none focus-visible:focus-ring-inset focus-visible:[--focus-ring:var(--common-white)]'
                    )}
                    to={data.project.route}>
                    <span
                        className={cn('block h-[30px] overflow-hidden', {
                            'w-[166px]': isExpanded,
                            'w-10': !isExpanded,
                        })}
                        data-testid='global_nav-logo'>
                        {data.project.icon}
                    </span>
                </AppLink>
            </div>
        </div>
    );
};

const MainNavListItem: FC<{
    /** Whether the main nav is in its expanded (wide) state; controls tooltip visibility */
    isExpanded: boolean;
    /** Typography variant applied to the item label */
    labelVariant: 'h5' | 'body1';
    /** The navigation item data to render, either a link, action, or subnav trigger */
    item: MainNavDataListItem;
}> = ({ isExpanded, item, labelVariant }) => {
    const location = useLocation();
    const [isSubNavOpen, setIsSubNavOpen] = useState(false);
    const navItemRef = useRef<HTMLLIElement>(null);
    const { control, icon, label, route, subNav, target, testId } = item;

    const isActiveRoute = isRouteActive(location.pathname, route);
    const isActiveSubNavRoute = subNav
        ? subNav.some((section) => section.items.some((item) => isRouteActive(location.pathname, item.path)))
        : false;
    const isSubNavVisible = subNav && isSubNavOpen;

    // Handles nav item text color and background with hover/active interactions
    const navItemContainerClasses = cn(
        'rounded flex items-center cursor-pointer focus-within:focus-ring-inset focus-within:[--focus-ring:var(--common-white)]',
        {
            'text-common-white bg-primary dark:text-primary-variant dark:bg-neutral-4':
                isActiveRoute || isActiveSubNavRoute,
            'group hover:text-common-white hover:bg-secondary dark:hover:text-[#7B78FD] dark:hover:bg-neutral-200':
                !isActiveRoute && !isActiveSubNavRoute,
        }
    );

    // Full width ensures that even clicking white space activates menu item
    const navItemClasses =
        'h-8 w-full min-w-0 overflow-hidden px-2 py-1 flex items-center gap-2 outline-none group-hover:cursor-pointer';

    const labelElement = (
        <span className='min-w-0 flex flex-1 items-center gap-2 overflow-hidden whitespace-nowrap'>
            <span className='flex size-6 shrink-0 items-center justify-center' data-testid='global_nav-item-label-icon'>
                {icon}
            </span>
            <Typography
                aria-label={label}
                className='block min-w-0 truncate !text-inherit'
                component='span'
                data-testid='global_nav-item-label-text'
                variant={labelVariant}>
                {label}
            </Typography>
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
            aria-current={isActiveRoute ? 'page' : undefined}
            className={navItemClasses}
            data-testid={testId}
            // PZ pages discard environment query params so all Zone Objects are counted
            // Some Objects do not have an environmentId (domain sid or tenant id)
            // As such, even using the "all" environments param does not capture everything
            discardQueryParams={route.includes(privilegeZonesPath)}
            target={target}
            to={route}>
            {labelElement}
            {target === '_blank' && isExpanded && (
                <FontAwesomeIcon className='shrink-0' icon={faExternalLink} size='sm' />
            )}
        </AppLink>
    ) : (
        <div
            aria-expanded={subNav ? isSubNavOpen : undefined}
            className={navItemClasses}
            data-testid={testId}
            onClick={onClick}
            onKeyDown={onKeyDown}
            role='button'
            tabIndex={0}>
            {labelElement}
            {control && isExpanded && <span className='ml-1 shrink-0'>{control}</span>}
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
    isExpanded: boolean;
}> = ({ image, isExpanded }) => {
    const { data: apiVersionResponse, isSuccess } = useApiVersion();
    const apiVersion = isSuccess && apiVersionResponse?.server_version;

    return (
        <div
            className={cn(
                'w-[264px] shrink-0 overflow-hidden border-primary pl-4 pr-2 text-left dark:border-neutral-4',
                'transition-all duration-300 ease-in',
                {
                    'h-[58px] py-2 border-t opacity-100 visible': isExpanded,
                    'h-0 py-0 border-t-0 opacity-0 invisible': !isExpanded,
                }
            )}
            data-testid='global_nav-footer'>
            {/* App version */}
            <Typography className='!text-[#BCB8E1]' data-testid='global_nav-version-number' variant='subtitle2'>
                BloodHound: {apiVersion}
            </Typography>

            {/* SpecterOps logo */}
            <Typography
                className='flex items-center gap-1 !text-inherit'
                component='div'
                data-testid='global_nav-powered-by'
                variant='caption'>
                Powered by
                <img
                    src={image.imageUrl}
                    alt={image.altText}
                    height={image.dimensions.height}
                    width={image.dimensions.width}
                    className={image.classes}
                />
            </Typography>
        </div>
    );
};

const MainNav: FC<{ mainNavData: MainNavData }> = ({ mainNavData }) => {
    const [isExpanded, setIsExpanded] = useNavExpanded();
    const isExtraLargeViewport = useMediaQuery(`(min-width: ${BREAKPOINTS.xl})`);
    const previousIsExtraLargeViewport = useRef(isExtraLargeViewport);
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

    useEffect(() => {
        if (previousIsExtraLargeViewport.current === isExtraLargeViewport) return;

        previousIsExtraLargeViewport.current = isExtraLargeViewport;
        setIsExpanded(isExtraLargeViewport);
    }, [isExtraLargeViewport, setIsExpanded]);

    const handleToggleNav = () => setIsExpanded(!isExpanded);

    return (
        <>
            <SkipLink href='#content-wrapper'>Skip to main content</SkipLink>
            {/* Nav expand/collapse button */}
            <IconButton
                aria-expanded={isExpanded}
                aria-label='Toggle Navigation'
                className={cn(
                    'h-8 w-8 p-2 border-none',
                    'text-common-white bg-primary dark:text-main dark:bg-neutral-5',
                    'hover:text-common-white hover:bg-secondary dark:hover:text-main dark:hover:bg-neutral-3',
                    'active:text-common-white active:bg-primary-variant dark:active:text-main dark:active:bg-neutral-2',
                    'focus-visible:text-common-white focus-visible:bg-secondary focus-visible:focus-ring-inset focus-visible:[--focus-ring:var(--common-white)]',
                    'dark:focus-visible:text-white dark:focus-visible:bg-neutral-3'
                )}
                onClick={handleToggleNav}
                size={16}
                tooltipPosition='right'
                // Negative right margin allows button to hover outside nav bar bounds
                triggerClassName={cn('absolute top-14 z-navToggle transition-all duration-300 ease-in', {
                    'rotate-180 left-[15.5rem]': isExpanded,
                    'left-10': !isExpanded,
                })}>
                <FontAwesomeIcon icon={faCaretRight} />
            </IconButton>

            <nav
                id='global-navigation'
                aria-label='Global navigation'
                tabIndex={-1}
                className={cn(
                    'flex flex-col flex-none font-medium z-nav print:hidden overflow-hidden',
                    'transition-all duration-300 ease-in',
                    'text-common-white bg-primary-variant dark:text-main dark:bg-neutral-300',
                    { 'basis-nav-width': !isExpanded, 'basis-nav-width-expanded': isExpanded }
                )}>
                {/* Bloodhound logo */}
                <MainNavLogo data={mainNavData.logo} isExpanded={isExpanded} />

                <div className='flex flex-col h-full overflow-x-hidden overflow-y-auto'>
                    {/* Nav menu top and bottom lists of items */}
                    <ul className='flex flex-col flex-grow gap-4 px-2 py-4' data-testid='global_nav-primary-list'>
                        {mainNavData.primaryList.map((item: MainNavDataListItem) => (
                            <MainNavListItem item={item} isExpanded={isExpanded} key={item.testId} labelVariant='h5' />
                        ))}
                    </ul>

                    <ul className='flex flex-col gap-2 px-2 py-4' data-testid='global_nav-secondary-list'>
                        {mainNavData.secondaryList.map((item: MainNavDataListItem) => (
                            <MainNavListItem
                                item={item}
                                isExpanded={isExpanded}
                                key={item.testId}
                                labelVariant='body1'
                            />
                        ))}
                    </ul>

                    <MainNavFooter image={mainNavData.logo.specterOps.image} isExpanded={isExpanded} />
                </div>
            </nav>
        </>
    );
};

export default MainNav;
