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
import { SeedTypeCypher } from 'js-client-library';
import { FC, useCallback, useContext, useMemo, useRef, useState } from 'react';
import { useQuery } from 'react-query';
import { graphSchema } from '../../../constants';
import { encodeCypherQuery } from '../../../hooks';
import { apiClient, cn } from '../../../utils';
import SelectorFormContext from '../Save/SelectorForm/SelectorFormContext';

const emptyFunction = () => {};

export const Cypher: FC<{
    preview?: boolean;
    initialInput?: string;
}> = ({ preview = true, initialInput = '' }) => {
    const [cypherQuery, setCypherQuery] = useState(initialInput);
    const [stalePreview, setStalePreview] = useState(false);
    const cypherEditorRef = useRef<CypherEditor | null>(null);

    const dispatch = useContext(SelectorFormContext).dispatch || emptyFunction;

    const kindsQuery = useQuery({
        queryKey: ['graph-kinds'],
        queryFn: ({ signal }) => apiClient.getKinds({ signal }).then((res) => res.data.data.kinds),
    });

    const schema = useCallback(() => graphSchema(kindsQuery.data), [kindsQuery.data]);

    const handleCypherSearch = useCallback(() => {
        if (preview) return;

        if (cypherQuery) {
            setStalePreview(false);
            dispatch({ type: 'set-seeds', seeds: [{ type: SeedTypeCypher, value: cypherQuery }] });
        }
    }, [cypherQuery, preview, dispatch]);

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
                                data-testid='zone-management_save_selector-form_update-results-button'
                                variant={'text'}
                                className={cn(
                                    'p-0 text-sm text-primary font-bold dark:text-secondary-variant-2 hover:no-underline',
                                    {
                                        'animate-pulse': stalePreview,
                                    }
                                )}
                                onClick={handleCypherSearch}>
                                Update Sample Results
                            </Button>
                        )}
                    </div>
                </div>
                {!preview && (
                    <p className='italic px-6 mt-2 text-sm'>
                        Note: The sample results from running this cypher search may include additional entities that
                        are not directly associated with the cypher query due to default selector expansion. In
                        contrast, 'View in Explore' will show only the entities that are directly associated with the
                        cypher query.
                    </p>
                )}
            </CardHeader>
            <CardContent className='px-6' data-testid='zone-management_cypher-container'>
                <div onClick={setFocusOnCypherEditor} className='flex-1' role='textbox'>
                    <CypherEditor
                        className={cn(
                            'flex flex-col border-solid border border-neutral-light-5 dark:border-neutral-dark-5 bg-white dark:bg-[#002b36] rounded-lg min-h-64 overflow-auto grow-1',
                            {
                                'bg-transparent [&_.cm-editor]:bg-transparent [&_.cm-editor_.cm-gutters]:bg-transparent [&_.cm-editor_.cm-gutters]:border-transparent dark:bg-transparent dark:[&_.cm-editor]:bg-transparent dark:[&_.cm-editor_.cm-gutters]:bg-transparent dark:[&_.cm-editor_.cm-gutters]:border-transparent':
                                    preview,
                            }
                        )}
                        ref={cypherEditorRef}
                        value={preview ? initialInput : cypherQuery}
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
