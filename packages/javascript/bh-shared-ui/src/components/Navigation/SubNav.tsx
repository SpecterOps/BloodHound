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
import { cn } from '../../utils';

const SubNavListItem: FC<{ children: ReactNode; route?: string }> = ({ children, route }) => {
    const location = useLocation();
    const isActiveRoute = route ? location.pathname.includes(route.replace(/\*/g, '')) : false;

    return (
        <li
            className={cn('w-full mx-2 mb-1 px-2 py-1 rounded', {
                'text-primary bg-neutral-light-4': isActiveRoute,
            })}>
            {children}
        </li>
    );
};

const SubNavListItemLink: FC<{ route: string; children: ReactNode }> = ({ route, children }) => {
    return (
        <RouterLink
            to={route as string}
            className={`h-8 min-h-8 w-full flex items-center gap-x-2 text-sm cursor-pointer whitespace-nowrap hover:text-secondary hover:underline dark:hover:text-secondary-variant-2`}>
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
        }[];
    }[];
}> = ({ sections }) => {
    return (
        <nav className='z-[nav - 1] w-subnav-width h-full flex flex-col gap-10 fixed top-0 left-nav-width bg-neutral-light-2 text-neutral-dark-0 pt-6 border-x border-solid border-neutral-light-5 dark:bg-neutral-dark-2 dark:text-neutral-light-1'>
            {sections.map((section, sectionIndex) => (
                <ul key={sectionIndex}>
                    <SubNavListItem>
                        <span className='font-medium decoration-none'>{section.title}</span>
                    </SubNavListItem>
                    {section.items.map((item, itemIndex) => (
                        <SubNavListItemLink key={itemIndex} route={item.path}>
                            <SubNavListItem route={item.path}>
                                <span>{item.label}</span>
                            </SubNavListItem>
                        </SubNavListItemLink>
                    ))}
                </ul>
            ))}
        </nav>
    );
};

export default SubNav;
