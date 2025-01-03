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

import { Browser, BrowserContext, Page } from '@playwright/test';
import PageManager from './pageManager.js';
import { invokeBrowser } from './browserManager.js';

export interface IFixture {
    browser: Browser;
    context: BrowserContext;
    pageManager: PageManager;
    page: Page;
}
// FixtureManager manages the state of the browser and page for each scenario
export default class FixtureManager implements IFixture {
    browser!: Browser;
    context!: BrowserContext;
    pageManager: PageManager;
    page!: Page;

    constructor() {
        this.pageManager = new PageManager();
    }

    get Fixture(): IFixture {
        return {
            browser: this.browser,
            context: this.context,
            pageManager: this.pageManager,
            page: this.pageManager.Page,
        };
    }

    async openBrowser() {
        this.browser = await invokeBrowser();
    }

    async openContext() {
        this.context = await this.browser.newContext();
    }

    async closeContext() {
        this.context.close();
    }

    async closeBrowser() {
        this.browser.close();
    }

    async newPage() {
        await this.pageManager.newPage(this.context);
    }
}
