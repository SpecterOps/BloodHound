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

import { useEffect } from 'react';
import ReactMarkdown, { type Components } from 'react-markdown';
import { Link, useLocation } from 'react-router-dom';
import remarkGfm from 'remark-gfm';
import { docImageUrls, docPages, docSectionAnchor, type DocPage } from '../docSource';

type HeadingTag = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';

const githubRepoUrl = 'https://github.com/SpecterOps/BloodHound';
// Placeholder origin that lets URL resolve repo-relative paths; it never appears in rendered output
const repoRootPlaceholderUrl = 'https://repo-root.invalid/';

const isExternalOrAnchor = (url: string) =>
    url.startsWith('#') || url.startsWith('//') || /^[a-z][a-z0-9+.-]*:/i.test(url);

const opensInNewTab = (url: string | undefined) => !!url && (url.startsWith('//') || /^https?:/i.test(url));

type RepoReference = { repoPath: string; resolved: URL };

// Resolves a relative markdown URL against the source file's repo location; null for external URLs and anchors
const resolveRepoReference = (url: string | undefined, sourcePath: string): RepoReference | null => {
    if (!url || isExternalOrAnchor(url)) return null;
    const resolved = new URL(url, `${repoRootPlaceholderUrl}${sourcePath}`);
    return { repoPath: decodeURIComponent(resolved.pathname.slice(1)), resolved };
};

export const DocPageContent = ({ docPage }: { docPage: DocPage }) => {
    const { hash } = useLocation();

    useEffect(() => {
        if (hash) document.getElementById(decodeURIComponent(hash.slice(1)))?.scrollIntoView();
    }, [docPage, hash]);

    const anchoredHeading = (Tag: HeadingTag): Components[HeadingTag] =>
        function AnchoredHeading({ node, children }) {
            return <Tag id={docSectionAnchor(docPage.sourcePath, node?.position?.start.line)}>{children}</Tag>;
        };

    const components: Components = {
        h1: anchoredHeading('h1'),
        h2: anchoredHeading('h2'),
        h3: anchoredHeading('h3'),
        h4: anchoredHeading('h4'),
        h5: anchoredHeading('h5'),
        h6: anchoredHeading('h6'),
        a: ({ href, title, children }) => {
            const reference = resolveRepoReference(href, docPage.sourcePath);
            if (!reference)
                return opensInNewTab(href) ? (
                    <a href={href} title={title} target='_blank' rel='noopener noreferrer'>
                        {children}
                    </a>
                ) : (
                    <a href={href} title={title}>
                        {children}
                    </a>
                );

            const { repoPath, resolved } = reference;
            const linkedPage = docPages.find((page) => page.sourcePath === repoPath);
            if (linkedPage)
                return (
                    <Link to={`${linkedPage.route}${resolved.hash}`} title={title}>
                        {children}
                    </Link>
                );

            const githubView = resolved.pathname.endsWith('/') ? 'tree' : 'blob';
            return (
                <a
                    href={`${githubRepoUrl}/${githubView}/main${resolved.pathname}${resolved.hash}`}
                    title={title}
                    target='_blank'
                    rel='noopener noreferrer'>
                    {children}
                </a>
            );
        },
        img: ({ src, alt, title }) => {
            const reference = resolveRepoReference(src, docPage.sourcePath);
            const bundledUrl = reference ? docImageUrls[reference.repoPath] : undefined;
            return <img src={bundledUrl ?? src} alt={alt} title={title} />;
        },
    };

    return (
        <article className='doc-page'>
            <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>
                {docPage.markdown}
            </ReactMarkdown>
        </article>
    );
};
