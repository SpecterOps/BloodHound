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

import { test } from 'bh-playwright-testing';
import { hideBySelector, restoreHidden } from 'bh-playwright-testing/axe';
import {
    installCreateUserTokenStub,
    installDeleteUserTokenStub,
    installMFAEnrollmentStub,
    installResetPasswordStub,
    installUserTokensStub,
} from 'bh-playwright-testing/stubs';

const password = process.env.A11Y_TEST_PASSWORD;

test.describe('WCAG A/AA violations - Profile', () => {
    test.beforeEach('setup', async ({ page, goAndWaitFor }) => {
        await goAndWaitFor('/ui/my-profile', page.getByRole('heading', { name: 'User Information' }));
    });

    test('Profile page', async ({ checkA11y }) => {
        await checkA11y();
    });

    test('API Key Management dialog - Create token', async ({ page, checkA11y }) => {
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
        await checkA11y({
            attachmentNamePrefix: 'create-token-empty',
            include: '[data-testid="user-token-management-dialog"]',
        });

        await page.getByRole('button', { name: 'Create Token' }).click();
        await page.getByRole('heading', { name: 'Create User Token' }).waitFor({ state: 'visible' });

        // Hide parent dialog in nested dialog scenario
        await hideBySelector(page, '[data-testid="user-token-management-dialog"]');

        // Create token form
        await checkA11y({
            attachmentNamePrefix: 'create-token-form',
            include: '[data-testid="create-user-token-dialog"]',
        });

        await page.getByRole('textbox', { name: 'Token Name' }).fill('Playwright Token');
        await page.getByRole('button', { name: 'Save' }).click();
        await page.getByText('Below is the new authentication token.').waitFor({ state: 'visible' });

        // Token list with new token
        await checkA11y({
            attachmentNamePrefix: 'create-token-success',
            include: '[data-testid="user-token-dialog"]',
        });
    });

    test('API Key Management dialog - Revoke token', async ({ page, checkA11y }) => {
        await installUserTokensStub(page);
        await installDeleteUserTokenStub(page);

        await page.getByRole('button', { name: 'API Key Management' }).click();
        await page.getByRole('button', { name: 'Revoke' }).waitFor({ state: 'visible' });

        // Dialogs can obscure page content causing false positives
        await hideBySelector(page, '#content-wrapper');

        // List of current tokens
        await checkA11y({
            include: '[data-testid="user-token-management-dialog"]',
            attachmentNamePrefix: 'revoke-token-list',
        });

        const revokeButton = page.getByRole('button', { name: 'Revoke' });
        await revokeButton.click();

        // Hide parent dialog in nested dialog scenario
        await hideBySelector(page, '[data-testid="user-token-management-dialog"]');

        await page.getByRole('heading', { name: 'Auth Token' }).waitFor({ state: 'visible' });

        // Revoke confirmation dialog for the stubbed token
        await checkA11y({
            attachmentNamePrefix: 'revoke-token-confirmation',
            include: '[data-testid="token-revoke-dialog"]',
        });

        // Successful confirmation returns to token list
    });

    test('Reset Password dialog', async ({ page, checkA11y }) => {
        // Stub the update password call so the flow can complete without changing the real password
        await installResetPasswordStub(page);

        // Open dialog
        const button = page.getByRole('button', { name: 'Reset Password' });
        await button.click();

        // Dialogs can obscure page content causeing false positives
        const hiddenContent = await hideBySelector(page, '#content-wrapper');

        // Password change form
        await checkA11y({
            attachmentNamePrefix: 'password-change-form',
            include: '[data-testid="password-dialog"]',
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
        await checkA11y({
            attachmentNamePrefix: 'password-change-validation',
            include: '[data-testid="password-dialog"]',
        });

        await restoreHidden(hiddenContent);

        // Fill out good password
        await newPassword.fill('#Ng%gLO$I(}!Iq8e5?uU');
        await newPasswordConfirm.fill('#Ng%gLO$I(}!Iq8e5?uU');
        await saveButton.click();

        // Password change success toast
        await checkA11y({
            attachmentNamePrefix: 'password-change-success',
            include: '.SnackbarContainer-root',
        });
    });

    test('Multi-Factor Authentication dialog', async ({ page, checkA11y }) => {
        await installMFAEnrollmentStub(page);

        // Open dialog
        const mfaToggle = page.getByRole('switch', { name: 'Disabled' });
        await mfaToggle.click();

        // Dialogs can obscure page content causing false positives
        await hideBySelector(page, '#content-wrapper');

        await page.getByText('To set up multi-factor authentication,').waitFor({ state: 'visible' });

        // Configure MFA dialog - input password
        await checkA11y({ attachmentNamePrefix: 'mfa-password', include: '[data-testid="enable-2fa-dialog"]' });

        const passwordInput = page.getByRole('textbox', { name: 'Password' });
        await passwordInput.fill(password);

        const nextButton = page.getByRole('button', { name: 'Next' });
        await nextButton.click();

        await page.getByRole('textbox', { name: 'One-Time Password' }).waitFor({ state: 'visible' });

        // Configure MFA dialog - input OTP
        await checkA11y({ attachmentNamePrefix: 'mfa-otp', include: '[data-testid="enable-2fa-dialog"]' });

        const otpInput = page.getByRole('textbox', { name: 'One-Time Password' });
        await otpInput.fill('123456');
        await nextButton.click();

        // Configure MFA dialog - success
        await checkA11y({ attachmentNamePrefix: 'mfa-success', include: '[data-testid="enable-2fa-dialog"]' });
    });
});
