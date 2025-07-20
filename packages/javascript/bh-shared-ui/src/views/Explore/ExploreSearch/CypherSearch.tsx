// Copyright 2023 Specter Ops, Inc.
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
import { faFolderOpen, faPlay, faQuestion, faSave } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useTheme } from '@mui/material';
import '@neo4j-cypher/codemirror/css/cypher-codemirror.css';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { useRef, useState } from 'react';
import { useQuery } from 'react-query';
import { graphSchema } from '../../../constants';
import { useCreateSavedQuery } from '../../../hooks';
import { useNotifications } from '../../../providers';
import { apiClient, cn } from '../../../utils';
import CommonSearches from './CommonSearches';
import SaveQueryDialog from './SaveQueryDialog';

type CypherSearchState = {
    cypherQuery: string;
    setCypherQuery: (query: string) => void;
    onCypherSearch: () => void;
    performSearch: (query?: string) => void;
};

type CypherSearchProps = {
    cypherSearchState: CypherSearchState;
    onCypherSearch: () => void;
};

const CypherSearch = ({ cypherSearchState, onCypherSearch }: CypherSearchProps) => {
    // Still using the MUI theme here to check for dark mode -- we need a better solution for this
    const theme = useTheme();

    const { cypherQuery, setCypherQuery, performSearch } = cypherSearchState;
    const createSavedQueryMutation = useCreateSavedQuery();

    const [showCommonQueries, setShowCommonQueries] = useState(false);
    const [showSaveQueryDialog, setShowSaveQueryDialog] = useState(false);

    const cypherEditorRef = useRef<CypherEditor | null>(null);

    const kindsQuery = useQuery({
        queryKey: ['graph-kinds'],
        queryFn: ({ signal }) => apiClient.getKinds({ signal }).then((res) => res.data.data.kinds),
    });

    const { addNotification } = useNotifications();

    const handleCypherSearch = () => {
        if (cypherQuery) {
            performSearch();
        }

        if (typeof onCypherSearch === 'function') {
            onCypherSearch();
        }
    };

    const handleSaveQuery = async (data: { name: string }) => {
        return createSavedQueryMutation.mutate(
            { name: data.name, query: cypherQuery },
            {
                onSuccess: () => {
                    setShowSaveQueryDialog(false);
                    addNotification(`${data.name} saved!`, 'userSavedQuery');
                },
            }
        );
    };

    const handleCloseSaveQueryDialog = () => {
        setShowSaveQueryDialog(false);
        createSavedQueryMutation.reset();
    };

    const setFocusOnCypherEditor = () => cypherEditorRef.current?.cypherEditor.focus();

    return (
        <>
            <div className='flex flex-col h-full'>
                <div className='flex gap-2 shrink-0'>
                    <Button
                        data-testid='explore_search-container_toggle-prebuilt-queries-button'
                        className='min-w-9 h-9 p-0'
                        variant={'secondary'}
                        onClick={() => {
                            setShowCommonQueries((v) => !v);
                        }}
                        aria-label='Show/Hide Saved Queries'>
                        <FontAwesomeIcon icon={faFolderOpen} />
                    </Button>

                    <div onClick={setFocusOnCypherEditor} className='flex-1' role='textbox'>
                        <CypherEditor
                            data-testid='explore_search-container_cypher-editor'
                            ref={cypherEditorRef}
                            className='flex grow flex-col border border-black/[.23] rounded bg-white dark:bg-[#002b36] min-h-24 max-h-24 overflow-auto [@media(min-height:720px)]:max-h-72 [&_.cm-tooltip]:max-w-lg'
                            value={cypherQuery}
                            onValueChanged={(val: string) => {
                                setCypherQuery(val);
                            }}
                            theme={theme.palette.mode}
                            onKeyDown={(e: any) => {
                                // if enter and shift key is pressed, execute cypher search
                                if (e.key === 'Enter' && e.shiftKey) {
                                    e.preventDefault();
                                    handleCypherSearch();
                                }
                            }}
                            schema={graphSchema(kindsQuery.data)}
                            lineWrapping
                            lint
                            placeholder='Cypher Query'
                            tooltipAbsolute={false}
                        />
                    </div>
                </div>

                <div className='flex gap-2 mt-2 justify-end shrink-0'>
                    <Button
                        variant='secondary'
                        onClick={() => {
                            setShowSaveQueryDialog(true);
                        }}
                        size={'small'}>
                        <div className='flex items-center'>
                            <FontAwesomeIcon icon={faSave} />
                            <p className='ml-2 text-base'>Save Query</p>
                        </div>
                    </Button>

                    <Button asChild variant='secondary' size={'small'}>
                        <a
                            href='https://bloodhound.specterops.io/analyze-data/bloodhound-gui/cypher-search'
                            rel='noreferrer'
                            target='_blank'>
                            <div className='flex items-center'>
                                <FontAwesomeIcon icon={faQuestion} />
                                <p className='ml-2 text-base'>Help</p>
                            </div>
                        </a>
                    </Button>

                    <Button onClick={() => handleCypherSearch()} size={'small'}>
                        <div className='flex items-center'>
                            <FontAwesomeIcon icon={faPlay} />
                            <p className='ml-2 text-base'>Run</p>
                        </div>
                    </Button>
                </div>

                <div className={cn('grow min-h-0', { hidden: !showCommonQueries })}>
                    <CommonSearches onSetCypherQuery={setCypherQuery} onPerformCypherSearch={performSearch} />
                </div>
            </div>
            <SaveQueryDialog
                open={showSaveQueryDialog}
                onClose={handleCloseSaveQueryDialog}
                onSave={handleSaveQuery}
                isLoading={createSavedQueryMutation.isLoading}
                error={createSavedQueryMutation.error}
            />
        </>
    );
};

export default CypherSearch;
