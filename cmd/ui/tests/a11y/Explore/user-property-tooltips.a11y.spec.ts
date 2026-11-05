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

import type { Page } from '@playwright/test';
import { expect, test } from 'bh-playwright-testing';

const mockNode = async (page: Page, id: number, kind: 'User' | 'Computer') => {
    await page.route(`**/api/v2/nodes/${id}*`, (route) =>
        route.fulfill({
            json: {
                data: {
                    node_id: id,
                    kinds: [{ name: kind, node_kind_id: 1 }],
                    properties: {
                        objectid: `playwright-tooltip-${id}`,
                        name: `TOOLTIP_${kind.toUpperCase()}@PLAYWRIGHT.LOCAL`,
                        admincount: true,
                        sensitive: false,
                        passwordnotreqd: false,
                    },
                },
            },
        })
    );
    await page.route(`**/api/v2/${kind === 'User' ? 'users' : 'computers'}/playwright-tooltip-${id}/*`, (route) =>
        route.fulfill({ json: { count: 0, data: [] } })
    );
};

test('User property help is discoverable by hover and keyboard without changing values or Computer labels', async ({
    page,
    checkA11y,
}) => {
    await mockNode(page, 42, 'User');
    await page.goto('/ui/explore?selectedItem=42');
    const panel = page.getByTestId('explore_entity-information-panel');
    const adminCount = panel.getByRole('button', { name: 'About Admin Count', exact: true });
    const sensitive = panel.getByRole('button', { name: 'About Marked Sensitive', exact: true });
    const password = panel.getByRole('button', { name: 'About Password Not Required', exact: true });
    const tooltip = page.getByRole('tooltip');
    await expect(panel.getByRole('button', { name: /^About / })).toHaveCount(3);

    await adminCount.hover();
    await expect(tooltip).toContainText('protected-account handling');
    await expect(tooltip).toContainText('Attribute Editor → adminCount');
    await expect(tooltip.getByRole('link')).toHaveCount(0);
    await expect(adminCount).toHaveAttribute('aria-describedby', (await tooltip.getAttribute('id'))!);
    await page.mouse.move(0, 0);
    await expect(tooltip).toBeHidden();
    await adminCount.locator('svg').hover();
    await expect(tooltip).toContainText('protected-account handling');
    await adminCount.press('Escape');
    await expect(tooltip).toBeHidden();

    await adminCount.press('Tab');
    await expect(sensitive).toBeFocused();
    await expect(tooltip).toContainText('prevents delegation');
    await expect(tooltip).toContainText('Properties → Account → Account options');
    await expect(sensitive).toHaveAttribute('aria-describedby', (await tooltip.getAttribute('id'))!);
    await sensitive.press('Escape');
    await expect(tooltip).toBeHidden();
    await expect(sensitive).toBeFocused();
    await sensitive.press('Tab');
    await expect(password).toBeFocused();
    await expect(tooltip).toContainText('does not establish that the password is blank');
    await expect(tooltip).toContainText('PASSWD_NOTREQD flag (0x20 / 32)');
    await expect(tooltip.getByRole('link')).toHaveCount(0);
    await checkA11y({ include: ['[data-testid="explore_entity-information-panel"]', '[role="tooltip"]'] });

    await password.press('Escape');
    const value = password.locator('../..').locator('[title="FALSE"]');
    await expect(value).toHaveText('FALSE');
    await value.hover();
    await expect(tooltip).toBeHidden();

    await mockNode(page, 43, 'Computer');
    await page.goto('/ui/explore?selectedItem=43');
    await expect(page.getByTestId('explore_entity-information-panel_header-text')).toHaveText(
        'TOOLTIP_COMPUTER@PLAYWRIGHT.LOCAL'
    );
    await expect(panel.getByText('Admin Count', { exact: true })).toBeVisible();
    await expect(panel.getByRole('button', { name: /^About / })).toHaveCount(0);
});
