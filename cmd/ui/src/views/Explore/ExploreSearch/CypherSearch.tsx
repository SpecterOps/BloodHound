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
import { Box, Collapse, Link, Typography } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import '@neo4j-cypher/codemirror/css/cypher-codemirror.css';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import {
    ActiveDirectoryKindProperties,
    ActiveDirectoryNodeKind,
    ActiveDirectoryRelationshipKind,
    AzureKindProperties,
    AzureNodeKind,
    AzureRelationshipKind,
    CommonKindProperties,
    searchbarActions,
    useCreateSavedQuery,
} from 'bh-shared-ui';
import { useState } from 'react';
import { addSnackbar } from 'src/ducks/global/actions';
import { useAppDispatch, useAppSelector } from 'src/store';
import CommonSearches from './CommonSearches';
import SaveQueryDialog from './SaveQueryDialog';

const useStyles = makeStyles((theme) => ({
    button: {
        minWidth: '35px',
        height: '35px',
    },
    iconButton: {
        padding: 0,
    },
    cypherEditor: {
        display: 'flex',
        flexGrow: 1,
        flexDirection: 'column',
        border: '1px solid',
        borderColor: 'rgba(0,0,0,.23)',
        borderRadius: theme.shape.borderRadius,
        backgroundColor: 'white',
        paddingTop: '5px',
        minHeight: '120px',
        // enables drag n drop resizing of editor
        resize: 'vertical',
        maxHeight: '500px',
    },
    cypherEditorDark: {
        display: 'flex',
        flexGrow: 1,
        flexDirection: 'column',
        border: '1px solid',
        borderColor: 'rgba(0,0,0,.23)',
        borderRadius: theme.shape.borderRadius,
        backgroundColor: '#002b36',
        paddingTop: '5px',
        minHeight: '120px',
        // enables drag n drop resizing of editor
        resize: 'vertical',
        maxHeight: '500px',
    },
}));

const schema = {
    labels: [
        ...Object.values(ActiveDirectoryNodeKind).map((nodeLabel) => `:${nodeLabel}`),
        ...Object.values(AzureNodeKind).map((nodeLabel) => `:${nodeLabel}`),
    ],
    relationshipTypes: [
        ...Object.values(ActiveDirectoryRelationshipKind).map((relationshipType) => `:${relationshipType}`),
        ...Object.values(AzureRelationshipKind).map((relationshipType) => `:${relationshipType}`),
    ],
    propertyKeys: [
        ...Object.values(CommonKindProperties),
        ...Object.values(ActiveDirectoryKindProperties),
        ...Object.values(AzureKindProperties),
    ],
};

const useCypherEditor = () => {
    const cypherQuery = useAppSelector((state) => state.search.cypher.searchTerm);

    const dispatch = useAppDispatch();

    const setCypherQuery = (query: string) => dispatch(searchbarActions.cypherQueryEdited(query));

    const performSearch = () => dispatch(searchbarActions.cypherSearch(cypherQuery));

    return {
        cypherQuery,
        setCypherQuery,
        performSearch,
    };
};

const CypherSearch = () => {
    const classes = useStyles();

    const { cypherQuery, setCypherQuery, performSearch } = useCypherEditor();
    const createSavedQueryMutation = useCreateSavedQuery();

    const [showCommonQueries, setShowCommonQueries] = useState(false);
    const [showEgg, setShowEgg] = useState(false);
    const [showSaveQueryDialog, setShowSaveQueryDialog] = useState(false);
    const dispatch = useAppDispatch();
    const darkMode = useAppSelector((state) => state.global.view.darkMode);

    const handleCypherSearch = () => {
        if (cypherQuery) {
            // Easter Egg Trigger
            if (cypherQuery.toLowerCase().includes('match (n) return n limit 5')) {
                setShowEgg(true);
            } else {
                setShowEgg(false);
            }
            performSearch();
        }
    };

    const handleSaveQuery = async (data: { name: string }) => {
        return createSavedQueryMutation.mutate(
            { name: data.name, query: cypherQuery },
            {
                onSuccess: () => {
                    setShowSaveQueryDialog(false);
                    dispatch(addSnackbar(`${data.name} saved!`, 'userSavedQuery'));
                },
            }
        );
    };

    const handleCloseSaveQueryDialog = () => {
        setShowSaveQueryDialog(false);
        createSavedQueryMutation.reset();
    };

    // work-around handler for user clicking within code-mirror <CypherEditor />
    const setFocusOnCypherEditor = () => {
        const input = document.querySelector('.cm-content') as HTMLElement;
        if (input) {
            input.focus();
        }
    };

    return (
        <>
            <Box display={'flex'} gap={1}>
                <Button
                    className={`${classes.button} ${classes.iconButton}`}
                    variant={'secondary'}
                    onClick={() => {
                        setShowCommonQueries((v) => !v);
                    }}
                    aria-label='Show/Hide Saved Queries'>
                    <FontAwesomeIcon icon={faFolderOpen} />
                </Button>

                <div onClick={setFocusOnCypherEditor} style={{ flex: 1 }} role='textbox'>
                    <CypherEditor
                        className={darkMode ? classes.cypherEditorDark : classes.cypherEditor}
                        value={cypherQuery}
                        onValueChanged={(val: string) => {
                            setCypherQuery(val);
                        }}
                        theme={darkMode ? 'dark' : 'light'}
                        onKeyDown={(e: any) => {
                            // if enter and shift key is pressed, execute cypher search
                            if (e.key === 'Enter' && e.shiftKey) {
                                e.preventDefault();
                                handleCypherSearch();
                            }
                        }}
                        schema={schema}
                        lineWrapping
                        lint
                        placeholder='Cypher Query'
                    />
                </div>
            </Box>

            <Box display={'flex'} gap={1} mt={1} justifyContent={'end'}>
                <Button
                    variant='secondary'
                    onClick={() => {
                        setShowSaveQueryDialog(true);
                    }}
                    size={'small'}>
                    <Box display={'flex'} alignItems={'center'}>
                        <FontAwesomeIcon icon={faSave} />
                        <Typography ml='8px'>Save Query</Typography>
                    </Box>
                </Button>

                <Button asChild variant='secondary' rel='noreferrer' size={'small'}>
                    <Link
                        href='https://support.bloodhoundenterprise.io/hc/en-us/articles/16721164740251'
                        target='_blank'>
                        <Box display={'flex'} alignItems={'center'}>
                            <FontAwesomeIcon icon={faQuestion} />
                            <Typography ml='8px'>Help</Typography>
                        </Box>
                    </Link>
                </Button>

                <Button onClick={() => handleCypherSearch()} size={'small'}>
                    <Box display={'flex'} alignItems={'center'}>
                        <FontAwesomeIcon icon={faPlay} />
                        <Typography ml='8px'>Run</Typography>
                    </Box>
                </Button>
            </Box>

            <Collapse in={showCommonQueries}>
                <CommonSearches />
            </Collapse>

            {showEgg && <EasterEgg />}

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

/*What is graphed will never die*/
const EasterEgg = () => {
    return (
        <div style={{ display: 'block', position: 'relative', bottom: 0, left: 0 }}>
            <img
                src={`${import.meta.env.BASE_URL}/img/logo-animated.gif`}
                alt='What is graphed will never die'
                style={{ width: '200px' }}
            />
        </div>
    );
};

export default CypherSearch;
