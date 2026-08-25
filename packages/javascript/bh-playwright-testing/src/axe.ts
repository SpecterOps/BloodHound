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
import type { ElementHandle, Locator, Page, TestInfo } from '@playwright/test';
import { expect, test as base } from '@playwright/test';
import type { AxeResults, NodeResult, Result } from 'axe-core';
import { installGraphHasDataStub } from './stubs/graphs/cypher';
import type { TestOptions } from './themes';

// Full list of supported tags here:
// https://www.deque.com/axe/core-documentation/api-documentation/#axecore-tags
export const WCAG_TAGS = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'] as const;

/**
 * Options for a single `checkA11y` scan. `include`/`exclude`/`disableRules` are applied to the
 * underlying `AxeBuilder` only when provided; `attachmentNamePrefix`/`maxNodesPerViolation` are
 * forwarded to `expectNoAccessibilityViolations` for report control.
 */
export type A11yScanOptions = {
    /**
     * Selector(s) to scope the scan to. Defaults to the `a11yDefaultInclude` option (unset in this
     * package; consumers prime it in their Playwright config — e.g. `'#content-wrapper'`). Pass
     * `null` to opt out of scoping entirely and scan the full page (e.g. a login screen).
     */
    include?: string | string[] | null;

    /** Selector(s) to omit from the scan (e.g. a known-noisy third-party embed). */
    exclude?: string | string[];

    /** Axe rule ids to turn off for this scan. */
    disableRules?: string | string[];

    /**
     * Namespaces the attachments (axe-results.json, screenshots, etc.) so multiple scans in one
     * test don't overwrite each other in the report.
     */
    attachmentNamePrefix?: string;

    /** Caps how many affected nodes per violation get a screenshot attachment. */
    maxNodesPerViolation?: number;
};

// Wait state for `goAndWaitFor`, mirroring Playwright's `Locator.waitFor` `state` option.
type WaitForState = 'attached' | 'detached' | 'visible' | 'hidden';

// The element `goAndWaitFor` waits on: either a ready-made `Locator` (locators are lazy, so it's
// safe to build one from `page` before navigating) or a factory that receives `page` — the factory
// form lets a caller avoid destructuring `page` when the locator is the only reason they'd need it.
type GoAndWaitForTarget = Locator | ((page: Page) => Locator);

type GoAndWaitForOptions = {
    /** Leave the global navigation drawer expanded instead of collapsing it after navigation. */
    keepNavOpen?: boolean;
    state?: WaitForState;
    timeout?: number;
};

type AxeFixtures = {
    makeAxeBuilder: () => AxeBuilder;

    /**
     * Describe-scoped defaults for `checkA11y`, declared as a Playwright option so a block can set
     * common values once via `test.use({ a11yDefaults: { include: '[role="dialog"]' } })`. Per-call
     * options passed to `checkA11y` win over these.
     */
    a11yDefaults: A11yScanOptions;

    /**
     * App-wide default scan scope used by `checkA11y` when neither a per-call option nor
     * `a11yDefaults` specifies `include`. Left `null` (full-page) here so the package makes no
     * assumption about a consumer's DOM; prime it per app in the Playwright config `use` block
     * (e.g. `a11yDefaultInclude: '#content-wrapper'`).
     */
    a11yDefaultInclude: string | string[] | null;

    /**
     * Accessible name of the button `goAndWaitFor` clicks to collapse the global navigation drawer.
     * Prime it per app in the Playwright config `use` block; defaults to `'Toggle Navigation'`.
     */
    navToggleName: string;

    /**
     * When `true`, the `page` fixture installs the shared cypher "has data" stub so `useGraphHasData`
     * resolves to "true" and the "No Data Available" upload dialog never settles open. Off by default
     * so the package stays inert; prime it per app in the Playwright config `use` block.
     */
    installGraphDataStub: boolean;

    /**
     * Runs a scoped axe scan and asserts there are no violations. Merges per-call options over the
     * describe-scoped `a11yDefaults`, falling back to the `a11yDefaultInclude` option for scope.
     *
     * @example
     * ```
     * test('empty', async ({ checkA11y }) => {
     *     await checkA11y();                          // uses a11yDefaults / a11yDefaultInclude
     * });
     * test('dialog', async ({ checkA11y }) => {
     *     await checkA11y({ include: '[role="dialog"]' });
     * });
     * ```
     */
    checkA11y: (options?: A11yScanOptions) => Promise<void>;

    /**
     * Navigates to `path`, collapses the global nav (unless `keepNavOpen`), then waits for `target`
     * to reach `state` (default `visible`). Collapses the `page.goto(...)` + `locator.waitFor(...)`
     * pair that opens most specs' `beforeEach`.
     *
     * @example
     * ```
     * test.beforeEach(async ({ goAndWaitFor, page }) => {
     *     await goAndWaitFor('/ui/download-collectors', page.getByRole('button', { name: 'Download SharpHound' }));
     * });
     * ```
     */
    goAndWaitFor: (path: string, target: GoAndWaitForTarget, options?: GoAndWaitForOptions) => Promise<void>;
};

// Composed `test` that adds:
//   - the `theme` worker-scoped option (consumed at config time via `TestOptions`)
//   - a `makeAxeBuilder` fixture preconfigured with the shared WCAG tag set
//   - `checkA11y` / `goAndWaitFor` helpers plus the options that parameterize them
//   - a `page` fixture that optionally installs the shared cypher "has data" stub
// Consumers that don't care about themes can ignore the option; it defaults to 'light'
// and has no runtime side effects.
export const test = base.extend<AxeFixtures, TestOptions>({
    // Injects window variable that may be checked by app at runtime
    // Allows BH to determine if it is run by Playwright to disable CSS transition animation
    context: async ({ context }, use) => {
        await context.addInitScript(() => {
            Object.defineProperty(window, '__APP_TEST_RUNTIME__', {
                value: {
                    type: 'accessibility',
                    runner: 'playwright',
                },
                configurable: false,
                writable: false,
            });
        });

        await use(context);
    },
    theme: ['light', { option: true, scope: 'worker' }],
    a11yDefaults: [{}, { option: true }],
    a11yDefaultInclude: [null, { option: true }],
    navToggleName: ['Toggle Navigation', { option: true }],
    installGraphDataStub: [false, { option: true }],
    makeAxeBuilder: async ({ page }, use, testInfo) => {
        testInfo.annotations.push({
            type: 'a11y-tags',
            description: WCAG_TAGS.join(', '),
        });

        await use(() => new AxeBuilder({ page }).withTags([...WCAG_TAGS]));
    },
    page: async ({ page, installGraphDataStub }, use) => {
        // Install the shared cypher "has data" stub before navigation when the consuming app opts in
        // (via the `installGraphDataStub` option). Individual tests can override the stub by
        // registering their own `page.route` for the cypher endpoint — Playwright runs handlers in
        // LIFO order, so a test-local handler wins for the cases it cares about.
        if (installGraphDataStub) {
            await installGraphHasDataStub(page);
        }

        await use(page);
    },
    checkA11y: async ({ page, makeAxeBuilder, a11yDefaults, a11yDefaultInclude }, use, testInfo) => {
        await use(async (options = {}) => {
            const {
                include = a11yDefaultInclude,
                exclude,
                disableRules,
                attachmentNamePrefix,
                maxNodesPerViolation,
            } = { ...a11yDefaults, ...options };

            let builder = makeAxeBuilder();
            // `include` may be `null` (explicit full-page scan), so guard on truthiness rather than
            // `undefined` — the destructuring default only fills in an omitted value.
            if (include) builder = builder.include(include);
            if (exclude) builder = builder.exclude(exclude);
            if (disableRules) builder = builder.disableRules(disableRules);

            const results = await builder.analyze();
            await expectNoAccessibilityViolations(testInfo, results, {
                page,
                attachmentNamePrefix,
                maxNodesPerViolation,
            });
        });
    },
    goAndWaitFor: async ({ page, navToggleName }, use) => {
        await use(async (path, target, options = {}) => {
            await page.goto(path);

            // Collapse the global navigation drawer by default so it doesn't overlap the content
            // being scanned. Pass `keepNavOpen` to leave it expanded (e.g. nav-focused specs).
            if (!options.keepNavOpen) {
                const expandedNav = page.getByRole('button', { name: navToggleName, expanded: true });
                // Best-effort: probe for the expanded toggle with a short timeout before clicking so
                // an already-collapsed drawer or a page without the app shell doesn't burn the full
                // click timeout. The brief wait still tolerates late hydration, and a missing toggle
                // never fails navigation.
                const navExpanded = await expandedNav
                    .waitFor({ state: 'visible', timeout: 1000 })
                    .then(() => true)
                    .catch(() => false);
                if (navExpanded) {
                    await expandedNav.click({ timeout: 3000 }).catch(() => {});
                }
            }

            const locator = typeof target === 'function' ? target(page) : target;
            await locator.waitFor({ state: options.state ?? 'visible', timeout: options.timeout });
        });
    },
});

export { expect };

// Combined Playwright options shape for a11y consumers. Pass to `defineConfig<A11yTestOptions>` so a
// config's `use` block can set the theme matrix option plus the a11y fixture options below.
export type A11yTestOptions = TestOptions &
    Pick<AxeFixtures, 'a11yDefaults' | 'a11yDefaultInclude' | 'navToggleName' | 'installGraphDataStub'>;

// Optional inputs that opt into per-node screenshot attachments. When `page` is provided,
// each violation's affected nodes are screenshot via Playwright and attached alongside the
// textual report so the Playwright/Allure report surfaces a visual indicator next to each
// violation. Without `page`, behavior is unchanged.
export type AttachAxeReportOptions = {
    attachmentNamePrefix?: string;
    maxNodesPerViolation?: number;
    page?: Page;
};

const DEFAULT_MAX_NODES_PER_VIOLATION = 5;

/**
 * Hide the background content. Dialogs often cover background content. When Axe assesses
 * the obscured content, it produces an `incomplete` result and oftent misses actual issues
 */
export async function hideBySelector(page: Page, selector: string) {
    return await page.addStyleTag({
        content: `
            ${selector} {
                visibility: hidden !important;
            }
        `,
    });
}

/**
 * Undo a previous `hideBySelector` call. Pass the `ElementHandle` returned by
 * `hideBySelector` to remove the injected `<style>` tag from the page, restoring
 * the previously hidden content.
 */
export async function restoreHidden(styleTag: ElementHandle<Node>) {
    await styleTag.evaluate((node) => node.parentNode?.removeChild(node));
    await styleTag.dispose();
}

export async function attachAxeReport(testInfo: TestInfo, results: AxeResults, opts: AttachAxeReportOptions = {}) {
    await testInfo.attach(prefixedAttachmentName('axe-results.json', opts.attachmentNamePrefix), {
        body: JSON.stringify(results, null, 2),
        contentType: 'application/json',
    });

    if (results.violations.length === 0) {
        return;
    }

    await testInfo.attach(prefixedAttachmentName('a11y-violations.md', opts.attachmentNamePrefix), {
        body: formatViolations(results.violations),
        contentType: 'text/markdown',
    });

    if (opts.page) {
        await attachViolationScreenshots(
            testInfo,
            opts.page,
            results.violations,
            opts.maxNodesPerViolation ?? DEFAULT_MAX_NODES_PER_VIOLATION,
            opts.attachmentNamePrefix
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
    maxNodesPerViolation: number,
    attachmentNamePrefix?: string
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

                await testInfo.attach(
                    prefixedAttachmentName(`a11y-${violation.id}-${nodeIndex + 1}.png`, attachmentNamePrefix),
                    {
                        body: screenshot,
                        contentType: 'image/png',
                    }
                );
            } catch {
                // Element may have detached, animated off-screen, or otherwise become
                // unscreenshottable between the axe scan and now. The textual report still
                // captures the failure; a missing screenshot shouldn't block the assertion.
            }
        }
    }
}

/**
 * Builds a Playwright attachment name, optionally namespaced for a specific axe scan.
 */
function prefixedAttachmentName(name: string, prefix?: string) {
    return prefix ? `${prefix}-${name}` : name;
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
