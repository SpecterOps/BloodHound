// Copyright 2024 Specter Ops, Inc.
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

import prisma from './client.js';

export class dbOPS {
    // delete test data in dev environment only
    async deleteUsers() {
        const baseURL = `${process.env.BASEURL}`.toLowerCase();
        if (`${process.env.ENV}` === 'dev' && (baseURL.includes('localhost') || baseURL.includes('127.0.0.1'))) {
            const deleteUsers = prisma.users.deleteMany();
            const deleteUserRoles = prisma.users_roles.deleteMany();
            await prisma.$transaction([deleteUserRoles, deleteUsers]);
        } else {
            console.log(
                `Skipping deletion test data, baseURL: ${process.env.BASEURL} does not target local environment`
            );
        }
    }
}
