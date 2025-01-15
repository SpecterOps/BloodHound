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

import { type Page } from '@playwright/test';

export default class LoginPage {
    constructor(private page: Page) {}

    async navigateToLoginPage() {
        this.page.goto(`${process.env.BASE_URL}/ui/login`, { waitUntil: 'domcontentloaded' });
    }

    async enterUserName(username: string) {
        return this.page.locator('#username').fill(username);
    }

    async enterEmail(email: string) {
        return this.page.locator('#username').fill(email);
    }

    async enterPassword(password: string) {
        this.page.locator('#password').fill(password);
    }

    async clickLoginButton() {
        this.page.getByRole('button', { name: 'LOGIN', exact: true }).click();
    }
}
