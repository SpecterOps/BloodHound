// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in wrißting, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import { hideMainContent } from 'bh-playwright-testing/axe';
import { installMFAEnrollmentStub } from 'bh-playwright-testing/stubs/mfa';
import { expectNoAccessibilityViolations, test } from '../../fixtures';

test.describe('WCAG A/AA violations - Profile', () => {
    test.beforeEach('setup', async ({ page }) => {
        await page.goto('/ui/my-profile');
    });

    test('Profile page', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.getByRole('heading', { name: 'User Information' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('API Key Management dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        // Open dialog
        const button = page.getByRole('button', { name: 'API Key Management' });
        await button.click();

        // Dialogs can obscure page content causeing false positives
        await hideMainContent(page);

        const results = await makeAxeBuilder().analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Reset Password dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        // Open dialog
        const button = page.getByRole('button', { name: 'Reset Password' });
        await button.click();

        // Dialogs can obscure page content causeing false positives
        await hideMainContent(page);

        const results = await makeAxeBuilder().include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('Multi-Factor Authentication dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        await installMFAEnrollmentStub(page);

        // Open dialog
        const mfaToggle = page.getByRole('switch', { name: 'Disabled' });
        await mfaToggle.click();

        // Dialogs can obscure page content causeing false positives
        await hideMainContent(page);

        let results = await makeAxeBuilder().include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page, attachmentNamePrefix: 'mfa-password' });

        const passwordInput = page.getByRole('textbox', { name: 'Password' });
        await passwordInput.fill(process.env.A11Y_TEST_PASSWORD);

        const nextButton = page.getByRole('button', { name: 'Next' });
        await nextButton.click();

        results = await makeAxeBuilder().include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page, attachmentNamePrefix: 'mfa-otp' });

        const otpInput = page.getByRole('textbox', { name: 'One-Time Password' });
        await otpInput.fill('123456');
        await nextButton.click();

        results = await makeAxeBuilder().include('[role="dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page, attachmentNamePrefix: 'mfa-success' });
    });
});
