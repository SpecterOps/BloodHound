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

import { AdministrationSection, Permission } from 'bh-shared-ui';

export const getAdminSubRoute = (route: string) => {
    const administrationRoute = '/administration/';
    return route.slice(administrationRoute.length);
};

export const getAdminFilteredSections = (
    sections: AdministrationSection[],
    checkAllPermissions: (permissions: Permission[]) => boolean
) => {
    // Checking these for now because the only route we are currently hiding is to the configuration page.
    // In practice, this will permit Administrators and Power User roles only.
    const hasAdminPermissions = checkAllPermissions([
        Permission.APP_READ_APPLICATION_CONFIGURATION,
        Permission.APP_WRITE_APPLICATION_CONFIGURATION,
    ]);
    return sections
        .map((section) => {
            const filteredItems = section.items.filter((item) => !item.adminOnly || hasAdminPermissions);
            return {
                ...section,
                items: filteredItems,
            };
        })
        .filter((section) => section.items.length !== 0);
};
