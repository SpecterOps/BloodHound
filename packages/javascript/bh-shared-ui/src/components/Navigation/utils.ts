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

import { AdministrationItem, AdministrationSection } from '../../types';

export const getSubRoute = (parentRoute: string, childRoute: string) => {
    return childRoute.slice(parentRoute.length);
};

/**
 * Takes a list of admin nav sections and returns a copy with "adminOnly" nav items removed and empty sections pruned
 */
export const filterAdminSections = (sections: AdministrationSection[]): AdministrationSection[] => {
    return sections
        .map((section) => {
            const filteredItems = section.items.filter((item) => !item.adminOnly);
            return {
                ...section,
                items: filteredItems,
            };
        })
        .filter((section) => section.items.length !== 0);
};

/**
 * Takes a nav section and returns a copy that has a nav item added if the title matches; the result is unmodified otherwise
 */
export const addItemToSection = (
    section: AdministrationSection,
    title: string,
    item: AdministrationItem
): AdministrationSection => {
    if (section.title === title) {
        return {
            ...section,
            items: [...section.items, item],
        };
    }
    return section;
};

/**
 * Flatten admin nav data to use when rendering each Route component
 */
export const flattenRoutes = (sections: AdministrationSection[]): AdministrationItem[] => {
    return sections.reduce<AdministrationItem[]>((acc, val) => acc.concat(val.items), []);
};
