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

import MiniSearch from 'minisearch';

export type DocPage = {
    sourcePath: string;
    route: string;
    markdown: string;
};

// A Doc Page split at its headings; the section before the first heading has an empty heading and anchor
export type DocSection = {
    id: string;
    sourcePath: string;
    route: string;
    line: number;
    heading: string;
    anchor: string;
    text: string;
};

export type DocSearchResult = DocSection & { snippet: string };

// Every glob below starts with this repo root, relative to this file. Vite requires the globs to be literal strings.
const repoRoot = '../../../';

// The markdown Doc Source include list
const includedFiles: Record<string, string> = import.meta.glob<string>(
    [
        '../../../rfc/**/*.md',
        '../../../DEVREADME.md',
        '../../../CONTRIBUTING.md',
        '../../../server/README.md',
        '../../../packages/javascript/*/README.md',
    ],
    { query: '?raw', import: 'default', eager: true }
);

// Images that included files can reference relatively, bundled by the Vite build. Scoped to the include list's folders.
const includedImages: Record<string, string> = import.meta.glob<string>(
    [
        '../../../rfc/**/*.{png,jpg,jpeg,gif,svg,webp}',
        '../../../*.{png,jpg,jpeg,gif,svg,webp}',
        '../../../server/*.{png,jpg,jpeg,gif,svg,webp}',
        '../../../packages/javascript/*/**/*.{png,jpg,jpeg,gif,svg,webp}',
        '!../../../packages/javascript/*/{src,node_modules,dist,storybook-static}/**',
    ],
    { query: '?url', import: 'default', eager: true }
);

export const docImageUrls: Record<string, string> = Object.fromEntries(
    Object.entries(includedImages).map(([importPath, url]) => [importPath.slice(repoRoot.length), url])
);

export const docsRoute = '/docs';

export const docPages: DocPage[] = Object.entries(includedFiles)
    .map(([importPath, markdown]) => {
        const sourcePath = importPath.slice(repoRoot.length);
        return { sourcePath, route: `${docsRoute}/${sourcePath.replace(/\.md$/, '')}`, markdown };
    })
    .sort((a, b) => a.sourcePath.localeCompare(b.sourcePath));

export const repoRootFolderLabel = 'Repo root';

export const docPageFolder = (sourcePath: string) => {
    const folderEnd = sourcePath.lastIndexOf('/');
    return folderEnd === -1 ? repoRootFolderLabel : sourcePath.slice(0, folderEnd);
};

export const docPageName = (sourcePath: string) => sourcePath.split('/').pop()!.replace(/\.md$/, '');

const atxHeading = /^ {0,3}#{1,6}\s+(.*?)(?:\s+#+)?\s*$/;
const codeFence = /^ {0,3}(```|~~~)/;

const plainHeading = (heading: string) => heading.replace(/\[([^\]]*)\]\([^)]*\)/g, '$1').replace(/[`*]/g, '');

// Matches GitHub's heading anchors, e.g. "7.4. Motivation & Goals" becomes "74-motivation--goals"
const headingSlug = (heading: string) =>
    heading
        .trim()
        .toLowerCase()
        .replace(/[^\p{L}\p{N}\s_-]/gu, '')
        .replace(/\s/g, '-');

const splitDocSections = (docPage: DocPage): DocSection[] => {
    const sectionDrafts = [{ line: 0, heading: '', anchor: '', lines: [] as string[] }];
    const slugCounts = new Map<string, number>();
    let inCodeFence = false;

    docPage.markdown.split(/\r?\n/).forEach((lineText, lineIndex) => {
        if (codeFence.test(lineText)) inCodeFence = !inCodeFence;
        const headingMatch = inCodeFence ? null : lineText.match(atxHeading);
        if (!headingMatch) {
            sectionDrafts[sectionDrafts.length - 1].lines.push(lineText);
            return;
        }

        const heading = plainHeading(headingMatch[1]);
        const slug = headingSlug(heading);
        const slugCount = slugCounts.get(slug) ?? 0;
        slugCounts.set(slug, slugCount + 1);
        sectionDrafts.push({
            line: lineIndex + 1,
            heading,
            anchor: slugCount === 0 ? slug : `${slug}-${slugCount}`,
            lines: [],
        });
    });

    return sectionDrafts
        .map(({ lines, ...sectionDraft }) => ({
            ...sectionDraft,
            id: `${docPage.sourcePath}:${sectionDraft.line}`,
            sourcePath: docPage.sourcePath,
            route: docPage.route,
            text: lines.join(' ').replace(/\s+/g, ' ').trim(),
        }))
        .filter((section) => section.heading || section.text);
};

const docSectionsById = new Map(
    docPages.flatMap(splitDocSections).map((section): [string, DocSection] => [section.id, section])
);

// The anchor of the heading that starts on the given 1-based line of a Doc Page's markdown
export const docSectionAnchor = (sourcePath: string, line: number | undefined) =>
    docSectionsById.get(`${sourcePath}:${line}`)?.anchor || undefined;

const docSearchIndex = new MiniSearch<DocSection>({
    fields: ['pageName', 'heading', 'text'],
    extractField: (section, field) =>
        field === 'pageName' ? docPageName(section.sourcePath) : String(section[field as keyof DocSection]),
    searchOptions: { prefix: true, fuzzy: 0.2, boost: { pageName: 2, heading: 2 } },
});
docSearchIndex.addAll(Array.from(docSectionsById.values()));

const snippetRadius = 80;

const docSnippet = (text: string, terms: string[]) => {
    const lowerText = text.toLowerCase();
    const matchIndex = Math.min(
        ...terms.map((term) => lowerText.indexOf(term)).filter((termIndex) => termIndex !== -1)
    );
    const start = Number.isFinite(matchIndex) ? Math.max(0, matchIndex - snippetRadius) : 0;
    const end = Math.min(text.length, start + snippetRadius * 2);
    return `${start > 0 ? '…' : ''}${text.slice(start, end)}${end < text.length ? '…' : ''}`;
};

export const searchDocs = (query: string): DocSearchResult[] =>
    docSearchIndex.search(query).map((result) => {
        const section = docSectionsById.get(result.id)!;
        return { ...section, snippet: docSnippet(section.text, result.terms) };
    });
