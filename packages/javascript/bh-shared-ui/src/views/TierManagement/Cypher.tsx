// Copyright 2025 Specter Ops, Inc.
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

import { Button } from '@bloodhoundenterprise/doodleui';
import '@neo4j-cypher/codemirror/css/cypher-codemirror.css';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { GraphNodes } from 'js-client-library';
import { FC, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useQuery } from 'react-query';
import { graphSchema } from '../../constants';
import { encodeCypherQuery } from '../../hooks';
import { apiClient } from '../../utils';

export const Cypher: FC<{
    preview?: boolean;
    initialInput?: string;
    setCypherSearchResults?: (nodes: GraphNodes | null) => void;
}> = ({ preview = true, initialInput = '', setCypherSearchResults }) => {
    const [cypherQuery, setCypherQuery] = useState(initialInput);
    const cypherEditorRef = useRef<CypherEditor | null>(null);

    const cypherUseQuery = useQuery({
        queryKey: ['tier-management', 'cypher'],
        queryFn: ({ signal }) => apiClient.cypherSearch(cypherQuery, { signal }).then((res) => res.data.data),
        retry: false,
        enabled: false,
    });

    const kindsQuery = useQuery({
        queryKey: ['graph-kinds'],
        queryFn: ({ signal }) => apiClient.getKinds({ signal }).then((res) => res.data.data.kinds),
    });

    useEffect(() => {
        if (!setCypherSearchResults) return;

        const result = cypherUseQuery.data ? cypherUseQuery.data.nodes : null;

        setCypherSearchResults(result);
    }, [cypherUseQuery.data, setCypherSearchResults]);

    const schema = useCallback(() => graphSchema(kindsQuery.data), [kindsQuery.data]);

    const handleCypherSearch = useCallback(() => {
        if (preview) return;
        if (cypherQuery) cypherUseQuery.refetch();
    }, [cypherUseQuery, cypherQuery, preview]);

    const onValueChanged = useCallback(
        (value: string) => {
            if (preview) return;
            setCypherQuery(value);
        },
        [preview, setCypherQuery]
    );

    const exploreUrl = useMemo(
        () => `/ui/explore?searchType=cypher&exploreSearchTab=cypher&cypherSearch=${encodeCypherQuery(cypherQuery)}`,
        [cypherQuery]
    );

    const setFocusOnCypherEditor = () => cypherEditorRef.current?.cypherEditor.focus();

    return (
        <section>
            <div className='h-full mt-4 rounded-lg'>
                <div className='flex justify-center rounded-lg bg-neutral-light-2 dark:bg-neutral-dark-2 min-h-80 min-w-[588px]'>
                    <div className='pt-6 pb-7 w-11/12 flex flex-col justify-around'>
                        <div className='flex justify-between items-center px-2'>
                            <header className='text-base font-bold'>
                                {preview ? 'Cypher Preview' : 'Cypher Search'}
                            </header>
                            <div className='flex gap-6'>
                                <Button variant={'text'} className='p-0 text-sm' asChild>
                                    <a href={exploreUrl} target='_blank' rel='noreferrer'>
                                        View in Explore
                                    </a>
                                </Button>
                                {!preview && (
                                    <Button
                                        variant={'text'}
                                        className='p-0 text-sm text-primary font-bold dark:text-secondary-variant-2 hover:no-underline'
                                        onClick={handleCypherSearch}>
                                        Run
                                    </Button>
                                )}
                            </div>
                        </div>
                        <div className='mt-4'>
                            <div onClick={setFocusOnCypherEditor} className='flex-1' role='textbox'>
                                <CypherEditor
                                    className='flex flex-col border-solid border border-black border-opacity-25 rounded-lg bg-white min-h-64 overflow-auto dark:bg-[#002b36] grow-1'
                                    ref={cypherEditorRef}
                                    value={cypherQuery}
                                    onValueChanged={onValueChanged}
                                    theme={document.documentElement.classList.contains('dark') ? 'dark' : 'light'}
                                    schema={schema()}
                                    readOnly={preview}
                                    lineWrapping
                                    lint
                                    placeholder='Cypher Query'
                                    tooltipAbsolute={false}
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </section>
    );
};
