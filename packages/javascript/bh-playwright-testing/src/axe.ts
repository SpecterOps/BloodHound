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

import AxeBuilder from '@axe-core/playwright';
import type { Page, TestInfo } from '@playwright/test';
import { expect, test as base } from '@playwright/test';
import type { AxeResults, NodeResult, Result } from 'axe-core';
import type { TestOptions } from './themes';

// Full list of supported tags here:
// https://www.deque.com/axe/core-documentation/api-documentation/#axecore-tags
export const WCAG_TAGS = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'] as const;

type AxeFixtures = {
    makeAxeBuilder: () => AxeBuilder;
};

// Composed `test` that adds:
//   - the `theme` worker-scoped option (consumed at config time via `TestOptions`)
//   - a `makeAxeBuilder` fixture preconfigured with the shared WCAG tag set
// Consumers that don't care about themes can ignore the option; it defaults to 'light'
// and has no runtime side effects.
export const test = base.extend<AxeFixtures, TestOptions>({
    theme: ['light', { option: true, scope: 'worker' }],
    makeAxeBuilder: async ({ page }, use, testInfo) => {
        testInfo.annotations.push({
            type: 'a11y-tags',
            description: WCAG_TAGS.join(', '),
        });

        await use(() => new AxeBuilder({ page }).withTags([...WCAG_TAGS]));
    },
});

export { expect };

// Optional inputs that opt into per-node screenshot attachments. When `page` is provided,
// each violation's affected nodes are screenshot via Playwright and attached alongside the
// textual report so the Playwright/Allure report surfaces a visual indicator next to each
// violation. Without `page`, behavior is unchanged.
export type AttachAxeReportOptions = {
    page?: Page;
    maxNodesPerViolation?: number;
};

const DEFAULT_MAX_NODES_PER_VIOLATION = 5;

export async function attachAxeReport(testInfo: TestInfo, results: AxeResults, opts: AttachAxeReportOptions = {}) {
    await testInfo.attach('axe-results.json', {
        body: JSON.stringify(results, null, 2),
        contentType: 'application/json',
    });

    if (results.violations.length === 0) {
        return;
    }

    await testInfo.attach('a11y-violations.md', {
        body: formatViolations(results.violations),
        contentType: 'text/markdown',
    });

    if (opts.page) {
        await attachViolationScreenshots(
            testInfo,
            opts.page,
            results.violations,
            opts.maxNodesPerViolation ?? DEFAULT_MAX_NODES_PER_VIOLATION
        );
    }
}

export async function expectNoAccessibilityViolations(
    testInfo: TestInfo,
    results: AxeResults,
    opts: AttachAxeReportOptions = {}
) {
    await attachAxeReport(testInfo, results, opts);

    expect(results.violations, formatViolations(results.violations)).toEqual([]);
}

function formatViolations(violations: Result[]) {
    if (violations.length === 0) {
        return 'No accessibility violations detected.';
    }

    return violations.map(formatViolation).join('\n\n---\n\n');
}

function formatViolation(violation: Result) {
    const affectedNodes = violation.nodes
        .slice(0, 10)
        .map((node) => {
            const target = node.target.join(' ');
            const failureSummary = node.failureSummary ? `\n  ${node.failureSummary}` : '';

            return `- \`${target}\`${failureSummary}`;
        })
        .join('\n');

    return `### ${violation.id} (${violation.impact ?? 'unknown impact'})
${violation.help}
${violation.helpUrl}

**Affected nodes:**
${affectedNodes}`;
}

async function attachViolationScreenshots(
    testInfo: TestInfo,
    page: Page,
    violations: Result[],
    maxNodesPerViolation: number
) {
    for (const violation of violations) {
        const nodes = violation.nodes.slice(0, maxNodesPerViolation);
        for (let nodeIndex = 0; nodeIndex < nodes.length; nodeIndex++) {
            const selector = selectorFromTarget(nodes[nodeIndex].target);
            if (selector === null) {
                // iframe / shadow-DOM target — Playwright can't resolve this from a single CSS
                // string. The textual report still describes the violation.
                continue;
            }

            try {
                const screenshot = await page
                    .locator(selector)
                    .first()
                    .screenshot({ animations: 'disabled', timeout: 2000 });

                await testInfo.attach(`a11y-${violation.id}-${nodeIndex + 1}.png`, {
                    body: screenshot,
                    contentType: 'image/png',
                });
            } catch {
                // Element may have detached, animated off-screen, or otherwise become
                // unscreenshottable between the axe scan and now. The textual report still
                // captures the failure; a missing screenshot shouldn't block the assertion.
            }
        }
    }
}

// Used in attachViolationScreenshots. When axe-core reports an accessibility violation,
// it states which DOM nodes are at fault. To take screenshots of those nodes, Playwright
// needs a CSS selector, like "button.submit", that it can pass to page.locator(...).
// This method convert axe's node description into plain selector strings.
//
// Axe's `node.target` is `(string | string[])[]`. A length > 1 entry indicates an iframe
// boundary crossing (each entry is the selector inside the corresponding frame); a `string[]`
// entry indicates shadow-DOM nesting. Both cases are skipped because Playwright requires a
// different API (frameLocator / `>>` engine) than a single CSS selector. For the common
// non-iframe, non-shadow case the target is `[string]` and the verbatim string is returned.
function selectorFromTarget(target: NodeResult['target']): string | null {
    if (target.length !== 1) return null;
    const first = target[0];
    if (typeof first !== 'string') return null;
    return first;
}
