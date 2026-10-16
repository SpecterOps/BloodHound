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

import { Input } from 'doodle-ui';
import { type FormEvent } from 'react';
import { createSearchParams, NavLink, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { DocPageContent } from '../components/DocPageContent';
import { DocSearchResults } from '../components/DocSearchResults';
import { PortalHeader } from '../components/PortalHeader';
import { docPageFolder, docPageName, docPages, docsRoute, type DocPage } from '../docSource';
import { NotFoundPage } from './NotFoundPage';

// Groups Doc Pages by their parent folder, keeping the sidebar to a folder level and a file level
const docPageGroups: [string, DocPage[]][] = Array.from(
    docPages
        .reduce((groups, docPage) => {
            const folder = docPageFolder(docPage.sourcePath);
            groups.set(folder, [...(groups.get(folder) ?? []), docPage]);
            return groups;
        }, new Map<string, DocPage[]>())
        .entries()
);

export const DocsPage = () => {
    const pagePath = useParams()['*'] ?? '';
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const query = searchParams.get('q')?.trim() ?? '';
    const currentPage = docPages.find((docPage) => docPage.route === `${docsRoute}/${pagePath}`);

    const submitSearch = (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        const submittedQuery = String(new FormData(event.currentTarget).get('q') ?? '').trim();
        navigate(submittedQuery ? `${docsRoute}?${createSearchParams({ q: submittedQuery })}` : docsRoute);
    };

    if (pagePath !== '' && !currentPage) return <NotFoundPage />;

    return (
        <div className='flex min-h-screen flex-col'>
            <div className='sticky top-0 z-10'>
                <PortalHeader />
            </div>
            <div className='flex flex-1 flex-col md:flex-row'>
                <nav
                    aria-label='Doc Pages'
                    className='max-h-[45vh] shrink-0 overflow-y-auto border-b border-dp-border bg-dp-surface p-5 md:sticky md:top-14 md:h-[calc(100vh-3.5rem)] md:max-h-none md:w-80 md:border-b-0 md:border-r'>
                    <form role='search' key={query} onSubmit={submitSearch} className='mb-6'>
                        <Input
                            type='search'
                            name='q'
                            variant='outlined'
                            aria-label='Search Repo Docs'
                            placeholder='Search Repo Docs'
                            defaultValue={query}
                        />
                    </form>
                    <ul className='flex flex-col gap-5'>
                        {docPageGroups.map(([folder, groupDocPages]) => (
                            <li key={folder}>
                                <span className='flex items-center gap-2.5 break-words pb-1.5 font-dp-display text-sm font-bold text-dp-text'>
                                    <span
                                        aria-hidden='true'
                                        className='h-3.5 w-3.5 shrink-0 rounded-full border-2 border-dp-accent'
                                    />
                                    {folder}
                                </span>
                                <ul aria-label={folder} className='doc-tree'>
                                    {groupDocPages.map((docPage) => (
                                        <li key={docPage.route}>
                                            <NavLink to={docPage.route} className='doc-tree-link'>
                                                {docPageName(docPage.sourcePath)}
                                            </NavLink>
                                        </li>
                                    ))}
                                </ul>
                            </li>
                        ))}
                    </ul>
                </nav>
                <main className='min-w-0 flex-1 px-6 py-10 md:px-12 lg:px-16'>
                    {currentPage ? (
                        <DocPageContent docPage={currentPage} />
                    ) : query ? (
                        <DocSearchResults query={query} />
                    ) : (
                        <p className='text-dp-text-muted'>Select a Doc Page from the sidebar.</p>
                    )}
                </main>
            </div>
        </div>
    );
};
