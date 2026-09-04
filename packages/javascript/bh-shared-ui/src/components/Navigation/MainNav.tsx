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
import { FC, useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { useLocation } from 'react-router-dom';
import { useApiVersion, useKeybindings, useNavExpanded } from '../../hooks';
import { privilegeZonesPath } from '../../routes';
import { cn, useAppNavigate } from '../../utils';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';
import { ConditionalTooltip } from '../ConditionalTooltip';
import { AppLink } from './AppLink';
import { SkipLink } from './SkipLink';
import SubNav from './SubNav';
import type { MainNavData, MainNavDataListItem, MainNavLogoDataObject, NavActionItem, NavLinkItem } from './types';

export type MainNavVariant = 'legacy' | 'refreshed';

export const MainNavLogo: FC<{
    data: MainNavLogoDataObject;
    isExpanded?: boolean;
    variant?: MainNavVariant;
}> = ({ data, isExpanded = true, variant = 'legacy' }) => {
    if (variant === 'refreshed') {
        return (
            <div className='h-[72px] flex-none border-b border-primary dark:border-neutral-4'>
                <div
                    className={cn('ml-2 mt-4 h-[30px] overflow-hidden', {
                        'w-[166px]': isExpanded,
                        'w-10': !isExpanded,
                    })}
                    data-testid='global_nav-home'>
                    <AppLink
                        aria-label='BloodHound home'
                        className='block h-[30px] w-[166px] focus-visible:focus-ring-inset focus-visible:[--focus-ring:var(--common-white)]'
                        to={data.project.route}>
                        {data.project.icon}
                    </AppLink>
                </div>
            </div>
        );
    }

    return (
        <div className='flex-none basis-10 m-2 mt-4 mb-6 overflow-hidden' data-testid='global_nav-home'>
            <AppLink to={data.project.route}>{data.project.icon}</AppLink>
        </div>
    );
};

const MainNavListItem: FC<{
    /** Whether the main nav is in its expanded (wide) state; controls tooltip visibility */
    isExpanded: boolean;
    labelVariant: 'h5' | 'body1';
    /** The navigation item data to render, either a link, action, or subnav trigger */
    item: MainNavDataListItem;
    variant: MainNavVariant;
}> = ({ isExpanded, item, labelVariant, variant }) => {
    const location = useLocation();
    const [isSubNavOpen, setIsSubNavOpen] = useState(false);
    const navItemRef = useRef<HTMLLIElement>(null);
    const { control, icon, label, route, subNav, target, testId } = item;

    const isActiveRoute = route ? location.pathname.includes(route.replace(/\*/g, '')) : false;
    const isActiveSubNavRoute = subNav
        ? subNav.some((section) => section.items.some((item) => location.pathname.includes(item.path)))
        : false;
    const isSubNavVisible = subNav && isSubNavOpen;

    const isActive = isActiveRoute || isActiveSubNavRoute;
    const navItemContainerClasses =
        variant === 'legacy'
            ? cn('text-xl rounded flex items-center cursor-pointer', {
                  'text-primary dark:text-[#8D8BF8] bg-neutral-4': isActive,
                  'group hover:text-primary-variant hover:dark:text-[#7B78FD] hover:bg-neutral-3 dark:hover:bg-[#1A1A1A]':
                      !isActive,
              })
            : cn(
                  'rounded flex items-center cursor-pointer focus-within:focus-ring-inset focus-within:[--focus-ring:var(--common-white)]',
                  {
                      'text-common-white bg-primary dark:text-[#8D8BF8] dark:bg-neutral-4': isActive,
                      'group hover:text-common-white hover:bg-secondary dark:hover:text-[#7B78FD] dark:hover:bg-[#1A1A1A]':
                          !isActive,
                  }
              );

    const navItemClasses = cn('w-full px-2 flex items-center group-hover:cursor-pointer', {
        'h-10 gap-x-2': variant === 'legacy',
        'h-8 min-w-0 overflow-hidden gap-2 py-1': variant === 'refreshed',
    });

    const labelElement =
        variant === 'legacy' ? (
            <span className='whitespace-nowrap flex items-center gap-x-2'>
                <span data-testid='global_nav-item-label-icon'>{icon}</span>
                <span data-testid='global_nav-item-label-text' aria-label={label}>
                    {label}
                </span>
            </span>
        ) : (
            <span className='min-w-0 flex flex-1 items-center gap-2 overflow-hidden whitespace-nowrap'>
                <span
                    className='flex size-6 shrink-0 items-center justify-center'
                    data-testid='global_nav-item-label-icon'>
                    {icon}
                </span>
                {isExpanded && (
                    <Typography
                        aria-label={label}
                        className='block min-w-0 truncate !text-inherit'
                        component='span'
                        data-testid='global_nav-item-label-text'
                        variant={labelVariant}>
                        {label}
                    </Typography>
                )}
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
            {...(variant === 'refreshed' ? { 'aria-label': label } : {})}
            to={route}>
            {labelElement}
            {target === '_blank' && (variant === 'legacy' || isExpanded) && (
                <FontAwesomeIcon
                    className={variant === 'refreshed' ? 'shrink-0' : undefined}
                    icon={faExternalLink}
                    size='sm'
                />
            )}
        </AppLink>
    ) : (
        <div
            className={navItemClasses}
            aria-expanded={subNav ? isSubNavOpen : undefined}
            {...(variant === 'refreshed' ? { 'aria-label': label } : {})}
            data-testid={testId}
            onClick={onClick}
            onKeyDown={onKeyDown}
            role='button'
            tabIndex={0}>
            {labelElement}
            {control && (variant === 'legacy' || isExpanded) && (
                <span className={cn('ml-1', variant === 'refreshed' && 'shrink-0')}>{control}</span>
            )}
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
                        variant={variant}
                    />,
                    navItemRef.current.closest('nav')!
                )}
        </>
    );
};

const MainNavFooter: FC<{
    image: MainNavLogoDataObject['specterOps']['image'];
    variant?: MainNavVariant;
}> = ({ image, variant = 'legacy' }) => {
    const { data: apiVersionResponse, isSuccess } = useApiVersion();
    const apiVersion = isSuccess && apiVersionResponse?.server_version;

    return variant === 'legacy' ? (
        <div className='py-3 text-xs'>
            <div className='flex flex-col w-[264px] items-center gap-2'>
                <div data-testid='global_nav-version-number'>BloodHound: {apiVersion}</div>

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
    ) : (
        <div className='h-[58px] w-[264px] border-t border-primary pl-4 pr-2 py-2 text-left dark:border-neutral-4'>
            <Typography className='!text-[#BCB8E1]' data-testid='global_nav-version-number' variant='subtitle2'>
                BloodHound: {apiVersion}
            </Typography>
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

const MainNav: FC<{ mainNavData: MainNavData; variant?: MainNavVariant }> = ({ mainNavData, variant = 'legacy' }) => {
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
            <SkipLink href='#content-wrapper'>Skip to main content</SkipLink>
            <IconButton
                aria-expanded={isExpanded}
                aria-label='Toggle Navigation'
                className={cn(
                    'absolute top-14 border-none z-navToggle',
                    'transition-all duration-300 ease-in',
                    variant === 'legacy'
                        ? 'min-h-6 min-w-6 w-5 p-1 text-main bg-neutral-4 dark:bg-neutral-5 hover:bg-[#B2B8BE] hover:text-main dark:hover:bg-neutral-3 dark:hover:text-main active:ring-0 active:bg-[#C0C6CB] dark:active:bg-neutral-2 dark:focus-visible:text-white'
                        : 'h-8 w-8 p-2 text-common-white bg-primary hover:text-common-white hover:bg-secondary active:text-common-white active:bg-primary-variant focus-visible:text-common-white focus-visible:bg-secondary focus-visible:focus-ring-inset focus-visible:[--focus-ring:var(--common-white)] dark:text-main dark:bg-neutral-5 dark:hover:text-main dark:hover:bg-neutral-3 dark:active:text-main dark:active:bg-neutral-2 dark:focus-visible:text-white dark:focus-visible:bg-neutral-3',
                    variant === 'legacy'
                        ? {
                              'rotate-180 left-[16.75rem]': isExpanded,
                              'left-[2.75rem]': !isExpanded,
                          }
                        : {
                              'rotate-180 left-[248px]': isExpanded,
                              'left-10': !isExpanded,
                          }
                )}
                size={16}
                onClick={handleToggleNav}>
                <FontAwesomeIcon icon={faCaretRight} />
            </IconButton>

            <nav
                id='global-navigation'
                aria-label='Global navigation'
                tabIndex={-1}
                className={cn(
                    'flex flex-col flex-none font-medium z-nav print:hidden overflow-hidden',
                    'transition-all duration-300 ease-in',
                    variant === 'legacy'
                        ? 'shadow-md bg-[#F2F2F2] dark:bg-[#1F1F1F]'
                        : 'text-common-white bg-primary-variant dark:text-main dark:bg-[#1F1F1F]',
                    variant === 'legacy'
                        ? { 'basis-nav-width': !isExpanded, 'basis-nav-width-expanded': isExpanded }
                        : {
                              'basis-14 w-14': !isExpanded,
                              'basis-[264px] w-[264px]': isExpanded,
                          }
                )}>
                <MainNavLogo data={mainNavData.logo} isExpanded={isExpanded} variant={variant} />

                <div
                    className={cn('flex flex-col h-full overflow-x-hidden overflow-y-auto', {
                        'mx-2': variant === 'legacy',
                    })}>
                    <ul
                        className={cn('flex flex-col flex-grow', variant === 'legacy' ? 'gap-2' : 'gap-4 px-2 py-4')}
                        data-testid='global_nav-primary-list'>
                        {mainNavData.primaryList.map((item: MainNavDataListItem) => (
                            <MainNavListItem
                                item={item}
                                isExpanded={isExpanded}
                                key={item.testId}
                                labelVariant='h5'
                                variant={variant}
                            />
                        ))}
                    </ul>

                    <ul
                        className={cn('flex flex-col gap-2', variant === 'legacy' ? 'mt-2' : 'px-2 py-4')}
                        data-testid='global_nav-secondary-list'>
                        {mainNavData.secondaryList.map((item: MainNavDataListItem) => (
                            <MainNavListItem
                                item={item}
                                isExpanded={isExpanded}
                                key={item.testId}
                                labelVariant='body1'
                                variant={variant}
                            />
                        ))}
                    </ul>

                    {(variant === 'legacy' || isExpanded) && (
                        <MainNavFooter image={mainNavData.logo.specterOps.image} variant={variant} />
                    )}
                </div>
            </nav>
        </>
    );
};

export default MainNav;
