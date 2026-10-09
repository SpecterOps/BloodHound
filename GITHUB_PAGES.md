<!--
Copyright 2026 Specter Ops, Inc.
SPDX-License-Identifier: Apache-2.0
-->

# GitHub Pages Layout

This repository publishes multiple static sites to a single GitHub Pages site,
served from the **`gh-pages`** branch (Pages source: *Deploy from a branch*).
A landing page at the root links to each sub-site.

```
https://specterops.github.io/BloodHound/            → landing page (index.html)
https://specterops.github.io/BloodHound/storybook/  → DoodleUI Storybook
https://specterops.github.io/BloodHound/allure/     → Allure test report
```

On the `gh-pages` branch this maps to:

```
gh-pages/
├── .nojekyll
├── index.html      # landing page
├── storybook/      # DoodleUI Storybook static build
└── allure/         # Allure report + history
```

## How it works

| Workflow | Trigger | Publishes to | Tool |
| --- | --- | --- | --- |
| `deploy-landing-page.yml` | push to `main` touching `.github/pages/index.html` (or manual) | `/` (root) | `peaceiris/actions-gh-pages` |
| `deploy-storybook.yml` | push to `main` touching `packages/javascript/doodle-ui/**` (or manual) | `/storybook/` | `peaceiris/actions-gh-pages` |
| `generate-allure-report.yml` | push to `main` | `/allure/` | `peaceiris/actions-gh-pages` |

Each workflow publishes with **`keep_files: true`** and its own
**`destination_dir`**, so it only ever updates its own subfolder and leaves the
others intact. The landing-page source of truth is `.github/pages/index.html`.

### Storybook subpath

Storybook is served under `/BloodHound/storybook/`, so `deploy-storybook.yml`
sets `STORYBOOK_BASE_PATH=/BloodHound/storybook/`. This is read by
`packages/javascript/doodle-ui/.storybook/main.ts` (in `viteFinal`, PRODUCTION
only) to set Vite's `base` so assets resolve. Local `yarn storybook` dev is
unaffected.

### Concurrency

All three workflows share `concurrency.group: gh-pages-deploy` with
`cancel-in-progress: false`, so pushes to `gh-pages` are serialized and cannot
clobber each other. Note the trade-off: if all three are queued from the same
event, GitHub allows one running + one pending per group and may drop the third.
In practice the path filters make simultaneous triggers rare.

### Caveat: `keep_files` is additive

Because `keep_files: true` never deletes, files removed from a sub-site (e.g. a
deleted story) linger until manually pruned. This is acceptable for docs/reports.

## One-time rollout

The steps below move the existing root Allure site into `/allure/` (preserving
history) and publish the landing page. **Run these once**, from the repo root.
They push to the `gh-pages` branch only; they do not touch `main`.

```bash
# 1. Relocate the current Allure site into an allure/ subfolder
git fetch origin gh-pages
git worktree add /tmp/bh-ghpages gh-pages
cd /tmp/bh-ghpages
mkdir -p allure
git ls-tree --name-only HEAD | grep -vE '^(allure|\.nojekyll)$' | while read -r item; do
  git mv "$item" allure/
done
touch .nojekyll
git add -A
git commit -m "chore(pages): relocate Allure report under /allure/ for shared landing page"
git push origin gh-pages
cd -
git worktree remove /tmp/bh-ghpages
```

Then publish the landing page and Storybook for the first time by manually
running the **Deploy Pages Landing Page** and **Deploy Storybook to GitHub
Pages** workflows from the Actions tab (both support `workflow_dispatch`), or by
merging a change that touches their respective paths.

No repository Pages **setting** needs to change — the source stays
`gh-pages` / root.

## Reverting to the previous configuration

The previous setup served the Allure report directly at the site root, with no
Storybook or landing page. To restore it:

1. **Remove the new files:**

```bash
git rm .github/workflows/deploy-storybook.yml \
       .github/workflows/deploy-landing-page.yml \
       .github/pages/index.html \
       GITHUB_PAGES.md
```

2. **Revert the edits** to these files (e.g. `git checkout <pre-session-commit> --`):
   - `.github/workflows/generate-allure-report.yml` — remove the `concurrency`
     block, `destination_dir: allure`, and `keep_files: true`; restore
     `gh_pages: gh-pages`.
   - `packages/javascript/doodle-ui/.storybook/main.ts` — remove the `viteFinal`
     block.

3. **Restore the `gh-pages` branch root.** With the reverted Allure workflow (no
   `keep_files`, no `destination_dir`), the next push to `main` republishes the
   report to the root and wipes the `storybook/`, `allure/`, and `index.html`
   entries automatically. To restore immediately instead of waiting:

```bash
git worktree add /tmp/bh-ghpages gh-pages
cd /tmp/bh-ghpages
rm -rf storybook index.html
git mv allure/* . 2>/dev/null || true
rmdir allure 2>/dev/null || true
git add -A
git commit -m "revert(pages): restore Allure report to gh-pages root"
git push origin gh-pages
cd -
git worktree remove /tmp/bh-ghpages
```
