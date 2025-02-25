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
import { FC, ReactNode } from 'react';
import { Link as RouterLink, useLocation } from 'react-router-dom';
import { cn, persistSearchParams } from '../../utils';

const SubNavListTitle: FC<{ children: ReactNode }> = ({ children }) => {
    return (
        <li className={'flex items-center mx-2 mb-1 px-2 text-neutral-dark-0 dark:text-neutral-light-1 font-medium'}>
            {children}
        </li>
    );
};

const SubNavListItem: FC<{ children: ReactNode; route?: string }> = ({ children, route }) => {
    const location = useLocation();
    const isActiveRoute = route ? location.pathname.includes(route.replace(/\*/g, '')) : false;

    return (
        <li
            className={cn('h-auto flex items-center mx-2 mb-1 px-2 rounded hover:underline', {
                'text-primary hover:text-primary bg-neutral-light-4': isActiveRoute,
                'text-neutral-dark-0 dark:text-neutral-light-1 hover:text-secondary dark:hover:text-secondary-variant-2':
                    !isActiveRoute,
            })}>
            {children}
        </li>
    );
};

const SubNavListItemLink: FC<{ route: string; persistentSearchParams?: string[]; children: ReactNode }> = ({
    route,
    children,
    persistentSearchParams,
}) => {
    const search = persistentSearchParams ? persistSearchParams(persistentSearchParams).toString() : undefined;
    return (
        <RouterLink
            to={{ pathname: route, search }}
            className={`h-7 min-h-7 w-full flex items-center gap-x-2 text-sm whitespace-nowrap`}>
            {children}
        </RouterLink>
    );
};

const SubNav: React.FC<{
    sections: {
        title: string;
        items: {
            path: string;
            label: string;
            persistentSearchParams?: string[];
        }[];
    }[];
}> = ({ sections }) => {
    return (
        <nav className='z-[nav - 1] w-subnav-width h-full flex flex-col gap-10 fixed top-0 left-nav-width bg-neutral-light-2 pt-6 border-x border-solid border-neutral-light-5 dark:bg-neutral-dark-2 overflow-x-hidden overflow-y-auto'>
            {sections.map((section, sectionIndex) => (
                <ul key={sectionIndex}>
                    <SubNavListTitle>
                        <span>{section.title}</span>
                    </SubNavListTitle>
                    {section.items.map((item, itemIndex) => (
                        <SubNavListItem key={itemIndex} route={item.path}>
                            <SubNavListItemLink route={item.path} persistentSearchParams={item.persistentSearchParams}>
                                <span>{item.label}</span>
                            </SubNavListItemLink>
                        </SubNavListItem>
                    ))}
                </ul>
            ))}
        </nav>
    );
};

export default SubNav;
