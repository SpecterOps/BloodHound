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
import { Button, Checkbox, Label } from '@bloodhoundenterprise/doodleui';
import { faChevronCircleRight, faSpinner } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useTheme } from '@mui/material';
import '@neo4j-cypher/codemirror/css/cypher-codemirror.css';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { UpdateUserQueryRequest } from 'js-client-library';
import { useCallback, useRef, useState } from 'react';
import { useQuery, useQueryClient } from 'react-query';
import { AppIcon } from '../../../components';
import { graphSchema } from '../../../constants';
import {
    decodeCypherQuery,
    useCreateSavedQuery,
    useExploreGraph,
    useExploreParams,
    useFeatureFlag,
    usePermissions,
    useQueryPermissions,
    useUpdateQueryPermissions,
    useUpdateSavedQuery,
} from '../../../hooks';
import { useNotifications } from '../../../providers';
import { Permission, apiClient, cn } from '../../../utils';
import { SavedQueriesProvider, useSavedQueriesContext } from '../providers';
import CommonSearches from './SavedQueries/CommonSearches';
import CypherSearchMessage, { CypherSearchMessageProps } from './SavedQueries/CypherSearchMessage';
import SaveQueryActionMenu from './SavedQueries/SaveQueryActionMenu';
import SaveQueryDialog from './SavedQueries/SaveQueryDialog';
import TagToZoneLabel from './SavedQueries/TagToZoneLabel';
import { CypherSearchState } from './types';

const CypherSearchInner = ({
    cypherSearchState,
    autoRun,
    setAutoRun,
}: {
    cypherSearchState: CypherSearchState;
    autoRun: boolean;
    setAutoRun: (autoRunQueries: boolean) => void;
}) => {
    const params = useExploreParams();

    const queryClient = useQueryClient();
    const { selectedQuery, saveAction, showSaveQueryDialog, setSelected, setSaveAction, setShowSaveQueryDialog } =
        useSavedQueriesContext();

    const { cypherQuery, setCypherQuery, performSearch } = cypherSearchState;

    const { data: featureFlagData, isLoading, isError } = useFeatureFlag('tier_management_engine');
    const privilegeZonesEnabled = !isLoading && !isError && featureFlagData?.enabled;

    const [showCommonQueries, setShowCommonQueries] = useState(false);
    const [messageState, setMessageState] = useState<CypherSearchMessageProps['messageState']>({
        showMessage: false,
        message: '',
    });

    const [sharedIds, setSharedIds] = useState<string[]>([]);
    const [isPublic, setIsPublic] = useState(false);
    const [saveUpdatePending, setSaveUpdatePending] = useState(false);

    // Still using the MUI theme here to check for dark mode -- we need a better solution for this
    const theme = useTheme();
    const createSavedQueryMutation = useCreateSavedQuery();
    const updateSavedQueryMutation = useUpdateSavedQuery();
    const updateQueryPermissionsMutation = useUpdateQueryPermissions();
    const kindsQuery = useQuery({
        queryKey: ['graph-kinds'],
        queryFn: ({ signal }) => apiClient.getKinds({ signal }).then((res) => res.data.data.kinds),
    });
    const { addNotification } = useNotifications();
    const { checkPermission } = usePermissions();

    const cypherEditorRef = useRef<CypherEditor | null>(null);
    const { data: permissions } = useQueryPermissions(selectedQuery?.id);

    const { isFetching: cypherSearchIsRunning, refetch } = useExploreGraph({
        onError: () => {
            setMessageState({
                showMessage: true,
                message: <span className='text-error'>Problem with query. Try again.</span>,
            });
        },
    });

    const cancelCypherQuery = () => {
        queryClient.cancelQueries({ queryKey: ['explore-graph-query'] });
    };

    const handleCypherSearch = () => {
        if (cypherQuery) {
            const cypherSearchInUrl = decodeCypherQuery(params.cypherSearch || '');
            const cypherInEditorMatchesCypherInUrl = cypherSearchState.cypherQuery === cypherSearchInUrl;

            if (cypherInEditorMatchesCypherInUrl) {
                // This reruns the previous search manually, since the user has not edited the query since last run.
                refetch();
            } else {
                // This injects fresh, user-edited cypher from editor into URL, triggering fresh query
                handleClearMessage();
                performSearch();
            }
        }
    };

    const handleSavedSearch = (query: string) => {
        if (autoRun) {
            handleClearMessage();
            performSearch(query);
        }
    };
    const handleToggleCommonQueries = () => {
        setShowCommonQueries((v) => !v);
    };
    const updateQueryPermissions = (id: number) => {
        if (permissions?.public && !isPublic && sharedIds.length) {
            const localSharedIds = [...sharedIds];
            updateQueryPermissionsMutation.mutate(
                {
                    id: id,
                    payload: {
                        user_ids: [],
                        public: isPublic,
                    },
                },
                {
                    onSettled: () => {
                        updateQueryPermissionsMutation.mutate({
                            id: id,
                            payload: {
                                user_ids: localSharedIds,
                                public: false,
                            },
                        });
                    },
                }
            );
        } else {
            updateQueryPermissionsMutation.mutate(
                {
                    id: id,
                    payload: {
                        user_ids: isPublic ? [] : sharedIds,
                        public: isPublic,
                    },
                },
                {
                    onSuccess: () => {
                        setSharedIds([]);
                        setIsPublic(false);
                    },
                }
            );
        }
    };

    const handleSaveQuery = async (data: { name: string; description: string; localCypherQuery: string }) => {
        setSaveUpdatePending(true);
        return createSavedQueryMutation.mutate(
            { name: data.name, description: data.description, query: data.localCypherQuery },
            {
                onSuccess: (res) => {
                    setShowSaveQueryDialog(false);
                    addNotification(`${data.name} saved!`, 'userSavedQuery');
                    performSearch(data.localCypherQuery);
                    setSelected({ query: data.localCypherQuery, id: res.id });
                    updateQueryPermissions(res.id);
                },
                onSettled: () => {
                    setSaveUpdatePending(false);
                },
            }
        );
    };

    const handleUpdateQuery = async (data: UpdateUserQueryRequest) => {
        setSaveUpdatePending(true);
        return updateSavedQueryMutation.mutate(
            { name: data.name, description: data.description, id: data.id, query: data.query },
            {
                onSuccess: (res) => {
                    setShowSaveQueryDialog(false);
                    setSelected({ query: data.query, id: data.id });
                    addNotification(`${data.name} updated!`, 'userSavedQuery');
                    performSearch(data.query);
                    updateQueryPermissions(res.id);
                },
                onSettled: () => {
                    setSaveUpdatePending(false);
                },
            }
        );
    };

    const handleClickSave = () => {
        if (selectedQuery) {
            if (selectedQuery.canEdit) {
                //save existing
                setSaveAction('edit');
                setShowSaveQueryDialog(true);
            } else {
                setMessageState({
                    showMessage: true,
                    message: 'You do not have permission to update this query, save as a new query instead',
                });
            }
        } else {
            //save new
            setSaveAction(undefined);
            setShowSaveQueryDialog(true);
        }
    };

    const handleClearMessage = useCallback(() => {
        setMessageState((prevState) => ({
            ...prevState,
            showMessage: false,
            message: '',
        }));
        // setTimeout(() => {
        //     setMessageState((prevState) => ({
        //         ...prevState,
        //         message: '',
        //     }));
        // }, 400);
    }, []);

    const handleCloseSaveQueryDialog = () => {
        setShowSaveQueryDialog(false);
        createSavedQueryMutation.reset();
        setSharedIds([]);
        setIsPublic(false);
    };

    const setFocusOnCypherEditor = () => cypherEditorRef.current?.cypherEditor.focus();

    const handleAutoRunQueryChange = (checked: boolean) => setAutoRun(checked);

    const handleSaveAs = () => {
        setSelected({ query: '', id: undefined });
        setSaveAction('save-as');
        setShowSaveQueryDialog(true);
    };

    const buttonText = cypherSearchIsRunning ? 'Abort' : 'Run';

    return (
        <>
            <div className='flex flex-col h-full'>
                {/* PRE BUILT SEARCHES SECTION */}
                <div className={cn('grow min-h-0 bg-[#f4f4f4] dark:bg-[#222222] p-2 py-0 rounded-lg mb-4')}>
                    <CommonSearches
                        onSetCypherQuery={setCypherQuery}
                        onPerformCypherSearch={handleSavedSearch}
                        onToggleCommonQueries={handleToggleCommonQueries}
                        showCommonQueries={showCommonQueries}
                    />
                </div>
                {/* CYPHER EDITOR SECTION */}
                <div className='bg-[#f4f4f4] dark:bg-[#222222] p-4 rounded-lg '>
                    <div className='flex items-center justify-between mb-2'>
                        <CypherSearchMessage messageState={messageState} clearMessage={handleClearMessage} />
                        <div className='flex items-center gap-4 whitespace-nowrap pr-2'>
                            <Checkbox
                                id='auto-run-selected-query'
                                checked={autoRun}
                                onCheckedChange={handleAutoRunQueryChange}
                            />
                            <Label htmlFor='auto-run-selected-query' className='font-normal cursor-pointer'>
                                Auto-run selected query
                            </Label>
                        </div>
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
                                onKeyDown={(e: React.KeyboardEvent<HTMLDivElement>) => {
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
                        {checkPermission(Permission.GRAPH_DB_WRITE) && privilegeZonesEnabled && (
                            <TagToZoneLabel cypherQuery={cypherSearchState.cypherQuery}></TagToZoneLabel>
                        )}
                        <Button
                            variant='secondary'
                            onClick={() => {
                                handleClickSave();
                            }}
                            size={'small'}
                            className='rounded-r-none'>
                            <div className='flex items-center'>
                                <p className='ml-2 text-base'>Save </p>
                            </div>
                        </Button>
                        <SaveQueryActionMenu saveAs={handleSaveAs} />

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

                        <Button
                            onClick={cypherSearchIsRunning ? cancelCypherQuery : handleCypherSearch}
                            size={'small'}
                            className={cn({
                                'bg-slate-600 hover:bg-slate-700': cypherSearchIsRunning,
                            })}>
                            <div className='flex items-center'>
                                <p className='text-base'>{buttonText}</p>
                                <FontAwesomeIcon
                                    size='lg'
                                    icon={cypherSearchIsRunning ? faSpinner : faChevronCircleRight}
                                    className={cn('ml-2', { 'animate-spin': cypherSearchIsRunning })}
                                />
                            </div>
                        </Button>
                        <div className='flex justify-center items-center'></div>
                    </div>
                </div>
            </div>

            <SaveQueryDialog
                open={showSaveQueryDialog}
                error={createSavedQueryMutation.error}
                cypherSearchState={cypherSearchState}
                sharedIds={sharedIds}
                isPublic={isPublic}
                saveAction={saveAction}
                saveUpdatePending={saveUpdatePending}
                onClose={handleCloseSaveQueryDialog}
                onSave={handleSaveQuery}
                onUpdate={handleUpdateQuery}
                setSharedIds={setSharedIds}
                setIsPublic={setIsPublic}
            />
        </>
    );
};

const CypherSearch = ({
    cypherSearchState,
    autoRun,
    setAutoRun,
}: {
    cypherSearchState: CypherSearchState;
    autoRun: boolean;
    setAutoRun: (autoRunQueries: boolean) => void;
}) => {
    return (
        <SavedQueriesProvider>
            <CypherSearchInner cypherSearchState={cypherSearchState} autoRun={autoRun} setAutoRun={setAutoRun} />
        </SavedQueriesProvider>
    );
};

export default CypherSearch;
