// Copyright 2026 Specter Ops, Inc.
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

import { expectNoAccessibilityViolations, test } from '../fixtures';

test.describe('Early Access Features page accessibility', () => {
    test('early access warning dialog has no detectable WCAG A/AA violations', async ({
        page,
        makeAxeBuilder,
    }, testInfo) => {
        await page.goto('/ui/administration/early-access-features');

        // This page will render a dialog warning upon mounting. Wait for the dialog to render before proceeding.
        await page.getByRole('heading', { name: 'Heads up!' }).waitFor({ state: 'visible' });

        // Dialog renders in portal, so div after content wrapper is assessed
        const results = await makeAxeBuilder().include('#content-wrapper + div').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });
});
