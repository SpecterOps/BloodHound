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

import { Given, When, Then } from "@cucumber/cucumber";
import { expect } from "@playwright/test";

Then('User is redirect to {string} page', async function (text: string) {
  await expect(this.fixture.page).toHaveURL(new RegExp(text +'$'));
});

Given('User enters invalid password', async function () {
  await this.fixture.page.locator("#password").fill("test1234");
});

When('User visits {string} page', async function (text: string) {
  await this.fixture.page.goto(`${process.env.BASEURL}/ui/${text}`);
});

When('User reloads the {string} page', async function (text: string) {
  await expect(this.fixture.page).toHaveURL(new RegExp(text+'$'));
  await this.fixture.page.reload({ waitUntil: "domcontentloaded" });
});

Then('User should be logged', async function () {
  await expect(this.fixture.page).toHaveURL(new RegExp('explore$'));
});

Then('User is redirect back to {string} page', async function (text: string) {
  await expect(this.fixture.page).toHaveURL(new RegExp(text + '$'));
});

Then('Page Displays Error Message', async function () {
  await expect(this.fixture.page.locator('#notistack-snackbar')).toBeVisible();
});
