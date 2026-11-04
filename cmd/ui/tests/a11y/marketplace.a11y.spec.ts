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

import { expect, expectNoAccessibilityViolations, test } from 'bh-playwright-testing';

test.describe('Marketplace page accessibility', () => {
    test('Marketplace content has no detectable WCAG A/AA violations', async ({ page, makeAxeBuilder }, testInfo) => {
        await page.goto('/ui/marketplace');
        await page.getByRole('heading', { name: 'Marketplace', exact: true }).waitFor({ state: 'visible' });
        await page.getByRole('region', { name: 'Community Extensions' }).waitFor({ state: 'visible' });

        const results = await makeAxeBuilder().include('#content-wrapper').analyze();
        await expectNoAccessibilityViolations(testInfo, results, { page });
    });

    test('development card labels and calls to action fit within a narrow viewport', async ({ page }) => {
        await page.setViewportSize({ width: 375, height: 812 });
        await page.goto('/ui/marketplace');
        await page.getByRole('button', { name: 'Toggle Navigation' }).click();

        const developmentCard = page.getByRole('article', { name: 'AWS Bedrock' });
        const category = developmentCard.getByText('Enterprise Extension');
        const signup = developmentCard.getByRole('link', { name: /Sign up to learn more AWS Bedrock/ });

        await developmentCard.scrollIntoViewIfNeeded();
        await expect(category).toBeVisible();
        await expect(signup).toBeVisible();

        const cardBounds = await developmentCard.boundingBox();
        const categoryBounds = await category.boundingBox();
        const signupBounds = await signup.boundingBox();
        const hasHorizontalOverflow = await developmentCard.evaluate((card) => card.scrollWidth > card.clientWidth);

        expect(cardBounds).not.toBeNull();
        expect(categoryBounds).not.toBeNull();
        expect(signupBounds).not.toBeNull();
        expect(hasHorizontalOverflow).toBe(false);

        for (const elementBounds of [categoryBounds, signupBounds]) {
            expect(elementBounds?.x).toBeGreaterThanOrEqual(cardBounds?.x ?? 0);
            expect((elementBounds?.x ?? 0) + (elementBounds?.width ?? 0)).toBeLessThanOrEqual(
                (cardBounds?.x ?? 0) + (cardBounds?.width ?? 0)
            );
            expect(elementBounds?.y).toBeGreaterThanOrEqual(cardBounds?.y ?? 0);
            expect((elementBounds?.y ?? 0) + (elementBounds?.height ?? 0)).toBeLessThanOrEqual(
                (cardBounds?.y ?? 0) + (cardBounds?.height ?? 0)
            );
        }
    });
});
