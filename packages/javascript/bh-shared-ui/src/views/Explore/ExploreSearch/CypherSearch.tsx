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
import { Checkbox, FormControlLabel, useTheme } from '@mui/material';
import '@neo4j-cypher/codemirror/css/cypher-codemirror.css';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { useEffect, useRef, useState } from 'react';
import { useQuery } from 'react-query';
import { AppIcon } from '../../../components';
import { graphSchema } from '../../../constants';
import { useCreateSavedQuery, useGetSelectedQuery } from '../../../hooks';
import { useNotifications } from '../../../providers';
import { apiClient, cn } from '../../../utils';
import CommonSearches from './CommonSearches';
import CypherSearchMessage from './CypherSearchMessage';
import SaveQueryDialog from './SaveQueryDialog';

type CypherSearchState = {
    cypherQuery: string;
    setCypherQuery: (query: string) => void;
    performSearch: (query?: string) => void;
};

const CypherSearch = ({ cypherSearchState }: { cypherSearchState: CypherSearchState }) => {
    // Still using the MUI theme here to check for dark mode -- we need a better solution for this
    const theme = useTheme();
    // const selectedQuery = useGetSelectedQuery();
    const [selected, setSelected] = useState('');
    const { cypherQuery, setCypherQuery, performSearch } = cypherSearchState;
    const createSavedQueryMutation = useCreateSavedQuery();

    const [showSaveQueryDialog, setShowSaveQueryDialog] = useState(false);
    const [showCommonQueries, setShowCommonQueries] = useState(false);

    const [autoRunQuery, setAutoRunQuery] = useState(true);
    const [messageState, setMessageState] = useState({
        showMessage: false,
        message: '',
    });

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
    };
    const handleSavedSearch = (query: any) => {
        if (autoRunQuery) {
            performSearch(query);
        }
    };
    useEffect(() => {
        setSelected(cypherQuery);
    }, [cypherQuery]);

    const handleSetSelected = (query: string) => {
        setSelected(query);
    };

    const selectedQuery: any = useGetSelectedQuery(selected);

    const handleToggleCommonQueries = () => {
        setShowCommonQueries((v) => !v);
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

    const handleClickSave = () => {
        console.log('handleClickSave');

        //FROM TICKET

        //IF QUERY SELECTED
        //  //  IF canEdit
        //  //  //  SAVE EXISTING WORKFLOW
        //  //  ELSE
        //  //  //  Display message:
        //  //  //“You do not have permission to update this     query, save as a new query instead”

        //IF NO QUERY SELECTED
        //  //  SAVE NEW WORKFLOW

        // if (!cypherQuery) {
        //     console.log('cypherQuery is empty - return');
        //     setMessageState({ showMessage: true, message: 'Add a Cypher Query' });
        //     return;
        // }

        if (selectedQuery) {
            console.log('QUERY EXISTS');
            if (selectedQuery.canEdit) {
                //save existing
                console.log('CAN EDIT - SAVE EXISTING WORKFLOW');
                setShowSaveQueryDialog(true);
            } else {
                console.log('CANNOT EDIT - SHOW MESSAGE');
                // setShowMessage(true);
                setMessageState({
                    showMessage: true,
                    message: 'You do not have permission to update this query, save as a new query instead',
                });
            }
        } else {
            //save new
            console.log('QUERY DOES NOT EXISTS - SAVE NEW WORKFLOW');
            setShowSaveQueryDialog(true);
        }
    };

    const handleClearMessage = () => {
        setMessageState((prevState) => ({
            ...prevState,
            showMessage: false,
        }));
        setTimeout(() => {
            setMessageState((prevState) => ({
                ...prevState,
                message: '',
            }));
        }, 400);
    };

    const handleCloseSaveQueryDialog = () => {
        setShowSaveQueryDialog(false);
        createSavedQueryMutation.reset();
    };

    const setFocusOnCypherEditor = () => cypherEditorRef.current?.cypherEditor.focus();

    const handleAutoRunQueryChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setAutoRunQuery(event.target.checked);
    };

    return (
        <>
            <div className='flex flex-col h-full'>
                {/* PRE BUILT SEARCHES SECTION */}
                <div className={cn('grow min-h-0 bg-[#f4f4f4] dark:bg-[#222222] p-2 py-0 rounded-lg mb-4')}>
                    <CommonSearches
                        onSetCypherQuery={setCypherQuery}
                        onPerformCypherSearch={handleSavedSearch}
                        onSetSelected={handleSetSelected}
                        onToggleCommonQueries={handleToggleCommonQueries}
                        showCommonQueries={showCommonQueries}
                        selected={selected}
                    />
                </div>
                {/* CYPHER EDITOR SECTION */}
                <div className='bg-[#f4f4f4] dark:bg-[#222222] p-4 rounded-lg '>
                    <div className='flex items-center justify-between mb-2'>
                        <CypherSearchMessage
                            messageState={messageState}
                            // showMessage={showMessage}
                            clearMessage={handleClearMessage}
                        />
                        <FormControlLabel
                            className='mr-0 whitespace-nowrap'
                            control={
                                <Checkbox
                                    checked={autoRunQuery}
                                    onChange={handleAutoRunQueryChange}
                                    inputProps={{ 'aria-label': 'controlled' }}
                                />
                            }
                            label='Auto-run selected query'
                        />
                    </div>

                    <div className='flex gap-2 shrink-0 '>
                        <div onClick={setFocusOnCypherEditor} className='flex-1' role='textbox'>
                            <CypherEditor
                                ref={cypherEditorRef}
                                className={cn(
                                    'flex grow flex-col border border-black/[.23] rounded bg-white dark:bg-[#002b36] min-h-24 max-h-24 overflow-auto [@media(min-height:720px)]:max-h-72 [&_.cm-tooltip]:max-w-lg',
                                    showCommonQueries && '[@media(min-height:720px)]:max-h-[20lvh]'
                                )}
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
                                // setShowSaveQueryDialog(true);
                                handleClickSave();
                            }}
                            size={'small'}
                            className='rounded-r-none'>
                            <div className='flex items-center'>
                                <p className='ml-2 text-base'>Save </p>
                            </div>
                        </Button>

                        <Button
                            variant='secondary'
                            onClick={() => {
                                setShowSaveQueryDialog(true);
                            }}
                            size={'small'}
                            className='rounded-l-none pl-2 -ml-1'>
                            <div className='flex items-center'>
                                <AppIcon.CaretDown size={10} />
                            </div>
                        </Button>

                        <Button asChild variant='secondary' size={'small'} className='px-1.5'>
                            <a
                                href='https://bloodhound.specterops.io/analyze-data/bloodhound-gui/cypher-search'
                                rel='noreferrer'
                                target='_blank'
                                className='group'>
                                <div>
                                    <AppIcon.Info size={24} />
                                </div>
                            </a>
                        </Button>

                        <Button onClick={() => handleCypherSearch()} size={'small'}>
                            <div className='flex items-center'>
                                <p className='text-base'>Run</p>
                            </div>
                        </Button>
                    </div>
                </div>
            </div>

            <SaveQueryDialog
                open={showSaveQueryDialog}
                onClose={handleCloseSaveQueryDialog}
                onSave={handleSaveQuery}
                isLoading={createSavedQueryMutation.isLoading}
                error={createSavedQueryMutation.error}
                cypherSearchState={cypherSearchState}
                selectedQuery={selectedQuery}
            />
        </>
    );
};

export default CypherSearch;
