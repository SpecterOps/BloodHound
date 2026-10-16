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

import { Link } from 'react-router-dom';
import { docPageFolder, docPageName, searchDocs } from '../docSource';

export const DocSearchResults = ({ query }: { query: string }) => {
    const results = searchDocs(query);

    return (
        <section className='max-w-[72ch]'>
            <h1 className='font-dp-display text-3xl font-extrabold leading-tight tracking-tight'>
                Search results for “{query}”
            </h1>
            {results.length === 0 ? (
                <p className='mt-4 text-dp-text-muted'>No Doc Pages match your search.</p>
            ) : (
                <>
                    <p className='mt-2 text-sm text-dp-text-muted'>
                        {results.length === 1 ? '1 result' : `${results.length} results`}
                    </p>
                    <ol
                        aria-label='Search results'
                        className='mt-8 divide-y divide-dp-border border-y border-dp-border'>
                        {results.map((result) => (
                            <li key={result.id} className='py-5'>
                                <span className='block text-xs text-dp-text-muted'>
                                    {docPageFolder(result.sourcePath)}
                                </span>
                                <Link
                                    to={result.anchor ? `${result.route}#${result.anchor}` : result.route}
                                    className='mt-1 inline-block font-dp-display text-lg font-bold text-dp-link underline-offset-2 hover:text-dp-link-hover hover:underline'>
                                    {docPageName(result.sourcePath)}
                                    {result.heading ? ` › ${result.heading}` : ''}
                                </Link>
                                {result.snippet && (
                                    <p className='mt-1 text-sm leading-6 text-dp-text-muted'>{result.snippet}</p>
                                )}
                            </li>
                        ))}
                    </ol>
                </>
            )}
        </section>
    );
};
