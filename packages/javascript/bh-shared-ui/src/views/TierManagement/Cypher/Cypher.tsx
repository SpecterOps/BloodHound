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

import { Button, Card, CardContent, CardHeader, CardTitle } from '@bloodhoundenterprise/doodleui';
import '@neo4j-cypher/codemirror/css/cypher-codemirror.css';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { AssetGroupTagNode, SeedTypeCypher } from 'js-client-library';
import { SelectorSeedRequest } from 'js-client-library/dist/requests';
import { FC, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useQuery } from 'react-query';
import { graphSchema } from '../../../constants';
import { encodeCypherQuery } from '../../../hooks';
import { apiClient, cn } from '../../../utils';
import './cypher.css';

export const Cypher: FC<{
    preview?: boolean;
    initialInput?: string;
    setSeedPreviewResults?: (nodes: AssetGroupTagNode[] | null) => void;
    setSeeds?: (seeds: SelectorSeedRequest[]) => void;
}> = ({ preview = true, initialInput = '', setSeedPreviewResults, setSeeds }) => {
    const [cypherQuery, setCypherQuery] = useState(initialInput);
    const [stalePreview, setStalePreview] = useState(false);
    const cypherEditorRef = useRef<CypherEditor | null>(null);

    const previewQuery = useQuery({
        queryKey: ['tier-management', 'preview-selectors', SeedTypeCypher, cypherQuery],
        queryFn: ({ signal }) =>
            apiClient
                .assetGroupTagsPreviewSelectors({ seeds: [{ type: SeedTypeCypher, value: cypherQuery }] }, { signal })
                .then((res) => res.data.data['members']),
        retry: false,
    });

    const kindsQuery = useQuery({
        queryKey: ['graph-kinds'],
        queryFn: ({ signal }) => apiClient.getKinds({ signal }).then((res) => res.data.data.kinds),
    });

    useEffect(() => {
        if (!setSeedPreviewResults) {
            return;
        }

        const result = previewQuery.data ? previewQuery.data : null;

        setSeedPreviewResults(result);
    }, [previewQuery.data, setSeedPreviewResults]);

    const schema = useCallback(() => graphSchema(kindsQuery.data), [kindsQuery.data]);

    const handleCypherSearch = useCallback(() => {
        if (preview) {
            return;
        }
        if (setSeeds) {
            setSeeds([{ type: SeedTypeCypher, value: cypherQuery }]);
        }
        if (cypherQuery) {
            previewQuery.refetch();
            setStalePreview(false);
        }
    }, [previewQuery, cypherQuery, preview, setSeeds]);

    const onValueChanged = useCallback(
        (value: string) => {
            if (preview) return;
            setCypherQuery(value);
            setStalePreview(true);
        },
        [preview, setCypherQuery]
    );

    const exploreUrl = useMemo(
        () => `/ui/explore?searchType=cypher&exploreSearchTab=cypher&cypherSearch=${encodeCypherQuery(cypherQuery)}`,
        [cypherQuery]
    );

    const setFocusOnCypherEditor = () => cypherEditorRef.current?.cypherEditor.focus();

    return (
        <Card>
            <CardHeader>
                <div className='flex justify-between items-center px-6 pt-3'>
                    <CardTitle>{preview ? 'Cypher Preview' : 'Cypher Search'}</CardTitle>
                    <div className='flex gap-6'>
                        <Button variant={'text'} className='p-0 text-sm' asChild>
                            <a href={exploreUrl} target='_blank' rel='noreferrer'>
                                View in Explore
                            </a>
                        </Button>
                        {!preview && (
                            <Button
                                variant={'text'}
                                className={cn(
                                    'p-0 text-sm text-primary font-bold dark:text-secondary-variant-2 hover:no-underline',
                                    {
                                        'animate-pulse': stalePreview,
                                    }
                                )}
                                onClick={handleCypherSearch}>
                                Run
                            </Button>
                        )}
                    </div>
                </div>
            </CardHeader>
            <CardContent className='px-6'>
                <div onClick={setFocusOnCypherEditor} className='flex-1' role='textbox'>
                    <CypherEditor
                        className={cn(
                            'flex flex-col border-solid border border-neutral-light-5 dark:border-neutral-dark-5 bg-white dark:bg-[#002b36] rounded-lg min-h-64 overflow-auto grow-1',
                            { 'bg-transparent dark:bg-transparent': preview }
                        )}
                        ref={cypherEditorRef}
                        value={cypherQuery}
                        onValueChanged={onValueChanged}
                        theme={document.documentElement.classList.contains('dark') ? 'dark' : 'light'}
                        schema={schema()}
                        readOnly={preview}
                        autofocus={false}
                        placeholder='Cypher Query'
                        tooltipAbsolute={false}
                        lineWrapping
                        lint
                    />
                </div>
            </CardContent>
        </Card>
    );
};
