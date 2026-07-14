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

import { hideBySelector, restoreHidden } from 'bh-playwright-testing/axe';
import {
    installCreateUserTokenStub,
    installDeleteUserTokenStub,
    installMFAEnrollmentStub,
    installResetPasswordStub,
    installUserTokensStub,
} from 'bh-playwright-testing/stubs';
import { expectNoAccessibilityViolations, test } from '../../fixtures';

const password = process.env.A11Y_TEST_PASSWORD;

test.describe('WCAG A/AA violations - Profile', () => {
    test.beforeEach('setup', async ({ page }) => {
        await page.goto('/ui/my-profile');
    });

    test('Profile page', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.getByRole('heading', { name: 'User Information' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('API Key Management dialog - Create token', async ({ page, makeAxeBuilder }, testInfo) => {
        // Render the empty token list and stub the create token call so the flow can complete
        await installUserTokensStub(page, { tokens: [] });
        await installCreateUserTokenStub(page);

        // Open dialog
        await page.getByRole('button', { name: 'API Key Management' }).click();

        // Dialogs can obscure page content causing false positives
        await hideBySelector(page, '#content-wrapper');

        await page.getByRole('heading', { name: 'Generate/Revoke API Tokens' }).waitFor({ state: 'visible' });
        await page.getByText('No tokens available').waitFor({ state: 'visible' });

        // Token management dialog with no tokens
        let results = await makeAxeBuilder().include('[data-testid="user-token-management-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page, attachmentNamePrefix: 'create-token-empty' });

        await page.getByRole('button', { name: 'Create Token' }).click();
        await page.getByRole('heading', { name: 'Create User Token' }).waitFor({ state: 'visible' });

        // Hide parent dialog in nested dialog scenario
        await hideBySelector(page, '[data-testid="user-token-management-dialog"]');

        // Create token form
        results = await makeAxeBuilder().include('[data-testid="create-user-token-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page, attachmentNamePrefix: 'create-token-form' });

        await page.getByRole('textbox', { name: 'Token Name' }).fill('Playwright Token');
        await page.getByRole('button', { name: 'Save' }).click();
        await page.getByText('Below is the new authentication token.').waitFor({ state: 'visible' });

        // Token list with new token
        results = await makeAxeBuilder().include('[data-testid="user-token-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, {
            page,
            attachmentNamePrefix: 'create-token-success',
        });
    });

    test('API Key Management dialog - Revoke token', async ({ page, makeAxeBuilder }, testInfo) => {
        await installUserTokensStub(page);
        await installDeleteUserTokenStub(page);

        await page.getByRole('button', { name: 'API Key Management' }).click();
        await page.getByRole('button', { name: 'Revoke' }).waitFor({ state: 'visible' });

        // Dialogs can obscure page content causing false positives
        await hideBySelector(page, '#content-wrapper');

        // List of current tokens
        let results = await makeAxeBuilder().include('[data-testid="user-token-management-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, {
            page,
            attachmentNamePrefix: 'revoke-token-list',
        });

        const revokeButton = page.getByRole('button', { name: 'Revoke' });
        await revokeButton.click();

        // Hide parent dialog in nested dialog scenario
        await hideBySelector(page, '[data-testid="user-token-management-dialog"]');

        await page.getByRole('heading', { name: 'Auth Token' }).waitFor({ state: 'visible' });

        // Revoke confirmation dialog for the stubbed token
        results = await makeAxeBuilder().include('[data-testid="token-revoke-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, {
            page,
            attachmentNamePrefix: 'revoke-token-confirmation',
        });

        // Successful confirmation returns to token list
    });

    test('Reset Password dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        // Stub the update password call so the flow can complete without changing the real password
        await installResetPasswordStub(page);

        // Open dialog
        const button = page.getByRole('button', { name: 'Reset Password' });
        await button.click();

        // Dialogs can obscure page content causeing false positives
        const hiddenContent = await hideBySelector(page, '#content-wrapper');

        // Password change form
        let results = await makeAxeBuilder().include('[data-testid="password-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, {
            page,
            attachmentNamePrefix: 'password-change-form',
        });

        // Fill out password change form to get failed validation state
        await page.getByRole('textbox', { name: 'Current Password' }).fill(password);

        const newPassword = page.getByRole('textbox', { name: 'New Password', exact: true });
        await newPassword.fill(password);

        const newPasswordConfirm = page.getByRole('textbox', { name: 'New Password Confirmation' });
        await newPasswordConfirm.fill(password);

        const saveButton = page.getByRole('button', { name: 'Save' });
        await saveButton.click();

        // Password change failed validation state
        results = await makeAxeBuilder().include('[data-testid="password-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, {
            page,
            attachmentNamePrefix: 'password-change-validation',
        });

        await restoreHidden(hiddenContent);

        // Fill out good password
        await newPassword.fill('#Ng%gLO$I(}!Iq8e5?uU');
        await newPasswordConfirm.fill('#Ng%gLO$I(}!Iq8e5?uU');
        await saveButton.click();

        // Password change success toast
        results = await makeAxeBuilder().include('.SnackbarContainer-root').analyze();
        await expectNoAccessibilityViolations(testInfo, results, {
            page,
            attachmentNamePrefix: 'password-change-success',
        });
    });

    test('Multi-Factor Authentication dialog', async ({ page, makeAxeBuilder }, testInfo) => {
        await installMFAEnrollmentStub(page);

        // Open dialog
        const mfaToggle = page.getByRole('switch', { name: 'Disabled' });
        await mfaToggle.click();

        // Dialogs can obscure page content causing false positives
        await hideBySelector(page, '#content-wrapper');

        await page.getByText('To set up multi-factor authentication,').waitFor({ state: 'visible' });

        // Configure MFA dialog - input password
        let results = await makeAxeBuilder().include('[data-testid="enable-2fa-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page, attachmentNamePrefix: 'mfa-password' });

        const passwordInput = page.getByRole('textbox', { name: 'Password' });
        await passwordInput.fill(password);

        const nextButton = page.getByRole('button', { name: 'Next' });
        await nextButton.click();

        await page.getByRole('textbox', { name: 'One-Time Password' }).waitFor({ state: 'visible' });

        // Configure MFA dialog - input OTP
        results = await makeAxeBuilder().include('[data-testid="enable-2fa-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page, attachmentNamePrefix: 'mfa-otp' });

        const otpInput = page.getByRole('textbox', { name: 'One-Time Password' });
        await otpInput.fill('123456');
        await nextButton.click();

        // Configure MFA dialog - success
        results = await makeAxeBuilder().include('[data-testid="enable-2fa-dialog"]').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page, attachmentNamePrefix: 'mfa-success' });
    });
});
