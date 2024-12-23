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

import { Given, When } from "@cucumber/cucumber";
import { User, IUserResult } from "../../../prisma/seed.js";

let newUser: Promise<IUserResult>

Given('Create a new user with {string} role', async function (roleType: string) {
    const user = new User();
    user.role = roleType
    newUser = user.create();
});

Given('Create a new user with {string} role with disabled status', async function (roleType: string) {
    const user = new User();
    user.role = roleType
    user.isDisabled = true
    newUser = user.create();
});

Given('User navigates to the login page', async function () {
    await this.fixture.page.goto(`${process.env.BASEURL}/ui/login`);
});

Given('User enters valid username', async function () {
    await this.fixture.page.locator("#username").fill((await newUser).principal_name);
});

Given('User enters valid email', async function () {
    await this.fixture.page.locator("#username").fill((await newUser).email_address);
  });
  
Given('User enters valid password', async function () {
    await this.fixture.page.locator("#password").fill((await newUser).uniquePassword);
});

When('User clicks on the login button', async function () {
    await this.fixture.page.getByRole('button', { name: "LOGIN", exact: true }).click();
});
  
