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

import { BrowserContext, Page } from '@playwright/test';

export default class PageManager {
    currentPage!: Page;

    get Page(): Page {
        return this.currentPage;
    }

    async newPage(context: BrowserContext): Promise<Page> {
        const page = await context.newPage();
        this.currentPage = page;
        return page;
    }
    async closePage(): Promise<void> {
        await this.currentPage.close();
    }
}
