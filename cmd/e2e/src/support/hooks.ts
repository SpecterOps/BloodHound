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

import { Before, After, BeforeAll, AfterAll, Status } from '@cucumber/cucumber';
import FixtureManager from './FixtureManager.js';
import PlaywrightWorld from '../tests/worlds/playwrightWorld.js';
import { loadEnvs } from '../helpers/env/env.js';
import { DbOPS } from '../../prisma/cleanup.js';

let fx: FixtureManager;

BeforeAll(async function () {
    // load environment variables
    loadEnvs();

    fx = new FixtureManager();
    await fx.openBrowser();
});

Before(async function (this: PlaywrightWorld) {
    // create new instance of fixture for each scenario
    await fx.openContext();
    await fx.newPage();
    this.fixture = fx.fixture;
});

After(async function ({ result, pickle }) {
    // capture screenshot for failed step
    if (result?.status === Status.FAILED) {
        const img = await this.fixture.page.screenshot({
            path: `./test-results/screenshots/+${pickle.name}`,
            type: 'png',
        });
        await this.attach(img, 'image/png');
    }

    // delete test users in dev environment
    const db = new DbOPS();
    db.deleteUsers();
});

AfterAll(async function () {
    await fx.closeBrowser();
});
