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
import { v4 as uuidv4 } from 'uuid';
import hashPassword from '../src/helpers/hashpassword.js';
import { faker } from '@faker-js/faker';
import { users } from '@prisma/client';

let uniquePassword: string;

// const is exported to ensure deletion of the test data.
export const qaEmailDomain = '@test.com';

export interface IUserResult extends users {
    email_address: string;
    uniquePassword: string;
    principal_name: string;
}

// create a unique user for each scenario
export class User {
    firstName: string;
    lastName: string;
    principalName: string;
    email: string;
    role: string;
    password?: string;
    eulaAccepted: boolean;
    isDisabled: boolean;
    isExpired: boolean;
    totpSecret: string;
    totpActivated: boolean;

    // initialize user properties as optional parameters
    constructor(
        firstName = faker.person.firstName(),
        lastName = faker.person.lastName(),
        principalName = `${firstName}-${lastName}`,
        email = `${firstName}${qaEmailDomain}`,
        role = 'Administrator',
        password = '',
        eulaAccepted = true,
        isDisabled = false,
        isExpired = false,
        totpSecret = '',
        totpActivated = false
    ) {
        this.firstName = firstName;
        this.lastName = lastName;
        this.principalName = principalName;
        this.email = email;
        this.password = password;
        this.role = role;
        this.eulaAccepted = eulaAccepted;
        this.isDisabled = isDisabled;
        this.isExpired = isExpired;
        this.totpSecret = totpSecret;
        this.totpActivated = totpActivated;
    }
    async create() {
        // sanity check when running against production environment
        // production tagged tests should not include any seeding data.
        if (process.env.ENV === 'production') {
            // throwing an _uncaught_ error exception to allow the process to be terminated accordingly
            throw new Error('Error: no seeding data in production environment');
        }

        // option to set the password in table driven tests
        if (this.password === '') {
            uniquePassword = faker.lorem.word({ length: { min: 5, max: 10 } });
        } else {
            uniquePassword = this.password as string;
        }

        const now = new Date();

        // find the matching role
        const role = await prisma.roles.findFirst({
            where: { name: this.role },
        });

        if (!role) {
            throw new Error(`Could not find the ${this.role} role`);
        }

        const newUser = await prisma.users.create({
            data: {
                id: uuidv4(),
                first_name: this.firstName,
                last_name: this.lastName,
                principal_name: this.principalName,
                email_address: this.email,
                eula_accepted: this.eulaAccepted,
                is_disabled: this.isDisabled,
                last_login: new Date(0),
                created_at: now,
                updated_at: now,
                users_roles: {
                    create: {
                        role_id: role.id,
                    },
                },
                auth_secrets: {
                    create: {
                        digest: await hashPassword(uniquePassword),
                        digest_method: 'argon2',
                        created_at: now,
                        updated_at: now,
                        expires_at: this.isExpired
                            ? new Date(0)
                            : new Date(new Date().setFullYear(new Date().getFullYear() + 1)),
                        totp_secret: this.totpSecret,
                        totp_activated: this.totpActivated,
                    },
                },
            },
        });

        return {
            ...newUser,
            uniquePassword,
        } as IUserResult;
    }
}
