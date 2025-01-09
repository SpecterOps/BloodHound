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
import { qaEmailDomain } from './seed.js';

export class DbOPS {
    qaEmail: string;

    constructor(qaEmail: string) {
      this.qaEmail = qaEmail;
    }
    // Delete test data in dev environment only
    async deleteUsers() {
        const baseURL = `${process.env.BASE_URL}`.toLowerCase();
        if (`${process.env.ENV}` === 'dev' && (baseURL?.includes('localhost') || baseURL?.includes('127.0.0.1'))) {
            const fetchAllTestUsers = await prisma.users.findMany({
                where: {
                    email_address: {
                        contains: this.qaEmail,
                    },
                },
            });
            for (const user of fetchAllTestUsers) {
                const deleteUserRoles = prisma.users_roles.deleteMany({
                    where: {
                        user_id: user.id,
                    },
                });
                const deleteUserSessions = prisma.user_sessions.deleteMany({
                    where: {
                        user_id: user.id,
                    },
                });
                const deleteAuthSecrets = prisma.auth_secrets.deleteMany({
                    where: {
                        user_id: user.id,
                    },
                });
                const deleteAuthTokens = prisma.auth_tokens.deleteMany({
                    where: {
                        user_id: user.id,
                    },
                });
                await prisma.$transaction([deleteUserRoles, deleteUserSessions, deleteAuthSecrets, deleteAuthTokens]);
            }
            const deleteUsers = prisma.users.deleteMany({
                where: {
                    email_address: {
                        contains: qaEmailDomain,
                    },
                },
            });
            await prisma.$transaction([deleteUsers]);
        } else {
            console.log(
                `Skipping deletion test data, baseURL: ${process.env.BASE_URL} does not target local environment`
            );
        }
    }
}
