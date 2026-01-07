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

import { FC, ReactNode, useMemo } from 'react';
import { useLocation } from 'react-router-dom';
import { useApiVersion, useIsMouseDragging, useKeybindings } from '../../hooks';
import { cn, useAppNavigate } from '../../utils';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';
import { AppLink } from './AppLink';
import { MainNavData, MainNavDataListItem, MainNavLogoDataObject } from './types';

const MainNavLogoTextImage: FC<{
    mainNavLogoData: MainNavLogoDataObject['specterOps'];
}> = ({ mainNavLogoData }) => {
    return (
        <img
            src={mainNavLogoData.image.imageUrl}
            alt={mainNavLogoData.image.altText}
            height={mainNavLogoData.image.dimensions.height}
            width={mainNavLogoData.image.dimensions.width}
            className={mainNavLogoData.image.classes}
        />
    );
};

const baseLinkContainerStyles = 'w-full min-h-10 overflow-hidden rounded';

export const MainNavLogo: FC<{ data: MainNavLogoDataObject; allowHover: boolean }> = (props) => {
    const { data, allowHover } = props;
    return (
        <div className={baseLinkContainerStyles} data-testid='global_nav-home'>
            <AppLink
                className={cn({
                    'cursor-pointer': allowHover,
                })}
                to={{ pathname: data.project.route }}>
                {data.project.icon}
            </AppLink>
        </div>
    );
};

const MainNavListItem: FC<{ children: ReactNode; route?: string; allowHover: boolean; onClick?: () => void }> = ({
    children,
    route,
    allowHover,
}) => {
    const location = useLocation();
    const isActiveRoute = route ? location.pathname.includes(route.replace(/\*/g, '')) : false;

    return (
        <li
            className={cn(
                baseLinkContainerStyles,
                'px-2 flex items-center text-neutral-dark-1 dark:text-neutral-light-1',
                {
                    'text-primary dark:text-primary dark:hover:text-primary bg-neutral-light-4 cursor-default':
                        isActiveRoute,
                },
                {
                    'hover:text-secondary dark:hover:text-secondary-variant-2': allowHover && !isActiveRoute,
                }
            )}>
            {children}
        </li>
    );
};

const MainNavItemAction: FC<{ onClick: () => void; children: ReactNode; allowHover: boolean; testId: string }> = ({
    onClick,
    children,
    allowHover,
    testId,
}) => {
    return (
        // Note: The w-full is to avoid the hover area to overflow out of the nav when its collapsed which created a flickering effect just outside the nav
        // Note: had to wrap in div to avoid error of button nesting in a button with the switch
        <div
            role='button'
            tabIndex={0}
            onKeyDown={adaptClickHandlerToKeyDown(onClick)}
            onClick={onClick}
            className={cn('h-10 w-auto flex items-center gap-x-2 hover:underline cursor-default', {
                'group-hover:w-full cursor-pointer': allowHover,
            })}
            data-testid={testId}>
            {children}
        </div>
    );
};

const MainNavItemLink: FC<{
    route: string;
    children: ReactNode;
    allowHover: boolean;
    testId: string;
}> = ({ route, children, allowHover, testId }) => {
    return (
        // Note: The w-full is to avoid the hover area to overflow out of the nav when its collapsed
        <AppLink
            to={{ pathname: route }}
            className={cn('h-10 w-auto flex items-center gap-x-2 hover:underline cursor-default', {
                'group-hover:w-full cursor-pointer': allowHover,
            })}
            data-testid={testId}>
            {children}
        </AppLink>
    );
};

const MainNavItemLabel: FC<{ icon: ReactNode; label: ReactNode | string }> = ({ icon, label }) => {
    return (
        // Note: The min-h here is to keep spacing between the logo and the list below.
        <>
            <span data-testid='global_nav-item-label-icon' className='flex'>
                {icon}
            </span>
            <span
                data-testid='global_nav-item-label-text'
                className='whitespace-nowrap min-h-10 font-medium text-xl transition-opacity duration-200 ease-in w-full opacity-100 flex items-center gap-x-5'>
                {label}
            </span>
        </>
    );
};

const MainNavVersionNumber: FC<{ allowHover: boolean }> = ({ allowHover }) => {
    const { data: apiVersionResponse, isSuccess } = useApiVersion();
    const apiVersion = isSuccess && apiVersionResponse?.server_version;

    return (
        // Note: The min-h allows for the version number to keep its position when the nav is scrollable
        <div className='relative w-full flex min-h-10 h-10' data-testid='global_nav-version-number'>
            <div
                className={cn(
                    'w-9 flex absolute top-3 left-3 duration-300 ease-in-out text-xs font-medium text-neutral-dark-1 dark:text-neutral-light-1',
                    {
                        'group-hover:w-auto group-hover:overflow-x-hidden group-hover:whitespace-nowrap group-hover:left-16':
                            allowHover,
                    }
                )}>
                <span
                    className={cn('opacity-0 hidden duration-300 ease-in-out', {
                        'group-hover:opacity-100 group-hover:block': allowHover,
                    })}>
                    BloodHound:&nbsp;{apiVersion}
                </span>
            </div>
        </div>
    );
};

const MainNavPoweredBy: FC<{ children: ReactNode; allowHover: boolean }> = ({ children, allowHover }) => {
    return (
        // Note: The min-h allows for the version number to keep its position when the nav is scrollable
        <div className='relative w-full flex min-h-10 h-10 overflow-x-hidden' data-testid='global_nav-powered-by'>
            <div
                className={cn(
                    'w-full flex absolute bottom-3 left-3 duration-300 ease-in-out text-xs whitespace-nowrap font-medium text-neutral-dark-1 dark:text-neutral-light-1',
                    {
                        'group-hover:left-12': allowHover,
                    }
                )}>
                <span
                    className={cn('opacity-0 hidden duration-300 ease-in-out ', {
                        'group-hover:opacity-100 group-hover:flex group-hover:items-center group-hover:gap-1':
                            allowHover,
                    })}>
                    powered by&nbsp;
                    {children}
                </span>
            </div>
        </div>
    );
};

const MainNav: FC<{ mainNavData: MainNavData }> = ({ mainNavData }) => {
    const { isMouseDragging } = useIsMouseDragging();
    const navigate = useAppNavigate();
    const allowHover = !isMouseDragging;

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
                'z-nav fixed top-0 px-2 left-0 h-full w-nav-width duration-300 ease-in flex flex-col items-center pt-4 shadow-sm bg-neutral-2 print:hidden group',
                {
                    'hover:w-nav-width-expanded hover:overflow-y-auto hover:overflow-x-hidden': allowHover,
                }
            )}>
            <MainNavLogo data={mainNavData.logo} allowHover={allowHover} />
            {/* Note: min height here is to keep the version number in bottom of nav */}
            <div className='h-full min-h-[665px] w-full flex flex-col justify-between'>
                <ul className='flex flex-col gap-4 mt-4' data-testid='global_nav-primary-list'>
                    {mainNavData.primaryList.map((listDataItem: MainNavDataListItem, itemIndex: number) => (
                        <MainNavListItem
                            key={itemIndex}
                            allowHover={!isMouseDragging}
                            route={listDataItem.route as string}>
                            {listDataItem.onClick && !listDataItem.route ? (
                                <MainNavItemAction
                                    onClick={listDataItem.onClick}
                                    allowHover={!isMouseDragging}
                                    testId={listDataItem.testId}>
                                    <MainNavItemLabel icon={listDataItem.icon} label={listDataItem.label} />
                                </MainNavItemAction>
                            ) : (
                                <MainNavItemLink
                                    route={listDataItem.route as string}
                                    allowHover={!isMouseDragging}
                                    testId={listDataItem.testId}>
                                    <MainNavItemLabel icon={listDataItem.icon} label={listDataItem.label} />
                                </MainNavItemLink>
                            )}
                        </MainNavListItem>
                    ))}
                </ul>
                <ul className='flex flex-col gap-4 mt-4' data-testid='global_nav-secondary-list'>
                    {mainNavData.secondaryList.map((listDataItem: MainNavDataListItem, itemIndex: number) =>
                        listDataItem.route ? (
                            <MainNavListItem
                                key={itemIndex}
                                route={listDataItem.route as string}
                                allowHover={allowHover}>
                                <MainNavItemLink
                                    route={listDataItem.route as string}
                                    allowHover={allowHover}
                                    testId={listDataItem.testId}>
                                    <MainNavItemLabel icon={listDataItem.icon} label={listDataItem.label} />
                                </MainNavItemLink>
                            </MainNavListItem>
                        ) : (
                            <MainNavListItem key={itemIndex} allowHover={allowHover}>
                                <MainNavItemAction
                                    onClick={listDataItem.functionHandler ?? (() => {})}
                                    allowHover={allowHover}
                                    testId={listDataItem.testId}>
                                    <MainNavItemLabel icon={listDataItem.icon} label={listDataItem.label} />
                                </MainNavItemAction>
                            </MainNavListItem>
                        )
                    )}
                </ul>
            </div>
            <MainNavVersionNumber allowHover={allowHover} />
            <MainNavPoweredBy allowHover={allowHover}>
                <MainNavLogoTextImage mainNavLogoData={mainNavData.logo.specterOps} />
            </MainNavPoweredBy>
        </nav>
    );
};

export default MainNav;
