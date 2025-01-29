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

import { FC, ReactNode, useState } from 'react';
import { Link as RouterLink, useLocation } from 'react-router-dom';
import { useApiVersion } from '../../hooks';
import { cn } from '../../utils';
import { MainNavData, MainNavDataListItem, MainNavLogoDataObject } from './types';

const MainNavLogoTextImage: FC<{
    mainNavLogoData: MainNavLogoDataObject['project'] | MainNavLogoDataObject['specterOps'];
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

const MainNavListItem: FC<{ children: ReactNode; route?: string }> = ({ children, route }) => {
    const location = useLocation();
    const isActiveRoute = route ? location.pathname.includes(route.replace(/\*/g, '')) : false;

    return (
        <li
            className={cn(
                'h-10 px-2 mx-2 flex items-center cursor-pointer rounded text-neutral-dark-0 dark:text-neutral-light-1 hover:text-secondary dark:hover:text-secondary-variant-2',
                {
                    'text-primary dark:text-primary dark:hover:text-primary bg-neutral-light-4': isActiveRoute,
                }
            )}>
            {children}
        </li>
    );
};

const MainNavItemAction: FC<{ onClick: () => void; children: ReactNode; isMenuExpanded: boolean }> = ({
    onClick,
    children,
    isMenuExpanded,
}) => {
    return (
        // Note: The w-full is to avoid the hover area to overflow out of the nav when its collapsed which created a flickering effect just outside the nav
        // Note: had to wrap in div to avoid error of button nesting in a button with the switch
        <div
            role='button'
            onClick={onClick}
            className={cn('h-10 w-auto absolute left-4 flex items-center gap-x-2 hover:underline', {
                'w-full': isMenuExpanded,
            })}>
            {children}
        </div>
    );
};

const MainNavItemLink: FC<{ route: string; children: ReactNode; isMenuExpanded: boolean }> = ({
    route,
    children,
    isMenuExpanded,
    ...rest
}) => {
    return (
        // Note: The w-full is to avoid the hover area to overflow out of the nav when its collapsed
        <RouterLink
            to={route as string}
            className={cn('h-10 w-auto absolute left-4 flex items-center gap-x-2 hover:underline', {
                'w-full': isMenuExpanded,
            })}
            {...rest}>
            {children}
        </RouterLink>
    );
};

const MainNavItemLabel: FC<{ icon: ReactNode; label: ReactNode | string; isMenuExpanded: boolean }> = ({
    icon,
    label,
    isMenuExpanded,
}) => {
    return (
        // Note: The min-h here is to keep spacing between the logo and the list below.
        <>
            <span data-testid='main-nav-item-label-icon' className='flex'>
                {icon}
            </span>
            <span
                data-testid='main-nav-item-label-text'
                className={cn(
                    'whitespace-nowrap min-h-10 font-medium text-xl opacity-0 hidden transition-opacity duration-200 ease-in',
                    { 'opacity-100 flex items-center gap-x-5': isMenuExpanded }
                )}>
                {label}
            </span>
        </>
    );
};

const MainNavVersionNumber: FC<{ isMenuExpanded: boolean }> = ({ isMenuExpanded }) => {
    const { data: apiVersionResponse, isSuccess } = useApiVersion();
    const apiVersion = isSuccess && apiVersionResponse?.server_version;

    return (
        // Note: The min-h allows for the version number to keep its position when the nav is scrollable
        <div className='relative w-full flex min-h-10 h-10 overflow-x-hidden' data-testid='main-nav-version-number'>
            <div
                className={cn(
                    'w-full flex absolute bottom-3 left-3 duration-300 ease-in-out text-xs whitespace-nowrap font-medium text-neutral-dark-0 dark:text-neutral-light-1',
                    { 'left-16': isMenuExpanded }
                )}>
                <span
                    className={cn('opacity-0 hidden duration-300 ease-in-out', {
                        'opacity-100 block': isMenuExpanded,
                    })}>
                    BloodHound:&nbsp;
                </span>
                <span
                    className={cn('', {
                        'max-w-9 overflow-x-hidden': !isMenuExpanded,
                    })}>
                    {apiVersion}
                </span>
            </div>
        </div>
    );
};

const MainNavPoweredBy: FC<{ isMenuExpanded: boolean; children: ReactNode }> = ({ isMenuExpanded, children }) => {
    return (
        // Note: The min-h allows for the version number to keep its position when the nav is scrollable
        <div className='relative w-full flex min-h-10 h-10 overflow-x-hidden' data-testid='main-nav-powered-by'>
            <div
                className={cn(
                    'w-full flex absolute bottom-3 left-3 duration-300 ease-in-out text-xs whitespace-nowrap font-medium text-neutral-dark-0 dark:text-neutral-light-1',
                    { 'left-12': isMenuExpanded }
                )}>
                <span
                    className={cn('opacity-0 hidden duration-300 ease-in-out', {
                        'opacity-100 flex items-center gap-1': isMenuExpanded,
                    })}>
                    powered by&nbsp;
                    {children}
                </span>
            </div>
        </div>
    );
};

const MainNav: FC<{ mainNavData: MainNavData }> = ({ mainNavData }) => {
    const [isMenuExpanded, setIsMenuExpanded] = useState(false);

    return (
        <nav
            className={cn(
                'z-nav fixed top-0 left-0 h-full w-nav-width duration-300 ease-in flex flex-col items-center pt-4 shadow-sm bg-neutral-light-2 dark:bg-neutral-dark-2 print:hidden',
                { 'w-nav-width-expanded overflow-y-auto overflow-x-hidden': isMenuExpanded }
            )}
            onMouseEnter={() => setIsMenuExpanded(true)}
            onMouseLeave={() => setIsMenuExpanded(false)}>
            <MainNavItemLink
                route={mainNavData.logo.project.route}
                isMenuExpanded={isMenuExpanded}
                data-testid='main-nav-logo'>
                <MainNavItemLabel
                    icon={mainNavData.logo.project.icon}
                    label={<MainNavLogoTextImage mainNavLogoData={mainNavData.logo.project} />}
                    isMenuExpanded={isMenuExpanded}
                />
            </MainNavItemLink>
            {/* Note: min height here is to keep the version number in bottom of nav */}
            <div className='h-full min-h-[700px] w-full flex flex-col justify-between mt-6'>
                <ul className='flex flex-col gap-6 mt-8' data-testid='main-nav-primary-list'>
                    {mainNavData.primaryList.map((listDataItem: MainNavDataListItem, itemIndex: number) => (
                        <MainNavListItem key={itemIndex} route={listDataItem.route as string}>
                            <MainNavItemLink route={listDataItem.route as string} isMenuExpanded={isMenuExpanded}>
                                <MainNavItemLabel
                                    icon={listDataItem.icon}
                                    label={listDataItem.label}
                                    isMenuExpanded={isMenuExpanded}
                                />
                            </MainNavItemLink>
                        </MainNavListItem>
                    ))}
                </ul>
                <ul className='flex flex-col gap-4 mt-16' data-testid='main-nav-secondary-list'>
                    {mainNavData.secondaryList.map((listDataItem: MainNavDataListItem, itemIndex: number) =>
                        listDataItem.route ? (
                            <MainNavListItem key={itemIndex} route={listDataItem.route as string}>
                                <MainNavItemLink route={listDataItem.route as string} isMenuExpanded={isMenuExpanded}>
                                    <MainNavItemLabel
                                        icon={listDataItem.icon}
                                        label={listDataItem.label}
                                        isMenuExpanded={isMenuExpanded}
                                    />
                                </MainNavItemLink>
                            </MainNavListItem>
                        ) : (
                            <MainNavListItem key={itemIndex}>
                                <MainNavItemAction
                                    onClick={(() => listDataItem.functionHandler as () => void)()}
                                    isMenuExpanded={isMenuExpanded}>
                                    <MainNavItemLabel
                                        icon={listDataItem.icon}
                                        label={listDataItem.label}
                                        isMenuExpanded={isMenuExpanded}
                                    />
                                </MainNavItemAction>
                            </MainNavListItem>
                        )
                    )}
                </ul>
            </div>
            <MainNavVersionNumber isMenuExpanded={isMenuExpanded} />
            <MainNavPoweredBy isMenuExpanded={isMenuExpanded}>
                <MainNavLogoTextImage mainNavLogoData={mainNavData.logo.specterOps} />
            </MainNavPoweredBy>
        </nav>
    );
};

export default MainNav;
