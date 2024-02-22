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

import { faFolderOpen, faPlay, faQuestion, faSave } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Button, Collapse, SvgIcon } from '@mui/material';
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
    useCreateSavedQuery,
} from 'bh-shared-ui';
import { useState } from 'react';
import { cypherQueryEdited, cypherSearch } from 'src/ducks/searchbar/actions';
import { useAppDispatch, useAppSelector } from 'src/store';
import CommonSearches from './CommonSearches';
import SaveQueryDialog from './SaveQueryDialog';
import { addSnackbar } from 'src/ducks/global/actions';

const useStyles = makeStyles((theme) => ({
    button: {
        minWidth: '35px',
        height: '35px',
        borderRadius: theme.shape.borderRadius,
        borderColor: 'rgba(0,0,0,0.23)',
        color: 'black',
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
        overflow: 'auto',
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

    const setCypherQuery = (query: string) => dispatch(cypherQueryEdited(query));

    const performSearch = () => dispatch(cypherSearch(cypherQuery));

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
                    onClick={() => {
                        setShowCommonQueries((v) => !v);
                    }}
                    variant='outlined'
                    aria-label='Show/Hide Saved Queries'>
                    <FontAwesomeIcon icon={faFolderOpen} />
                </Button>

                <div onClick={setFocusOnCypherEditor} style={{ flex: 1 }} role='textbox'>
                    <CypherEditor
                        className={classes.cypherEditor}
                        value={cypherQuery}
                        onValueChanged={(val: string) => {
                            setCypherQuery(val);
                        }}
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
                        placeholder='Cypher Search'
                    />
                </div>
            </Box>

            <Box display={'flex'} gap={1} mt={1} justifyContent={'end'}>
                <Button
                    className={classes.button}
                    onClick={() => {
                        setShowSaveQueryDialog(true);
                    }}
                    variant='outlined'
                    startIcon={
                        <SvgIcon>
                            <FontAwesomeIcon icon={faSave} />
                        </SvgIcon>
                    }>
                    Save Query
                </Button>

                <Button
                    href='https://support.bloodhoundenterprise.io/hc/en-us/articles/16721164740251'
                    target='_blank'
                    rel='noreferrer'
                    variant='outlined'
                    className={classes.button}
                    startIcon={
                        <SvgIcon>
                            <FontAwesomeIcon icon={faQuestion} />
                        </SvgIcon>
                    }>
                    Help
                </Button>

                <Button
                    className={classes.button}
                    onClick={() => handleCypherSearch()}
                    variant='outlined'
                    startIcon={
                        <SvgIcon>
                            <FontAwesomeIcon icon={faPlay} />
                        </SvgIcon>
                    }>
                    Search
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
