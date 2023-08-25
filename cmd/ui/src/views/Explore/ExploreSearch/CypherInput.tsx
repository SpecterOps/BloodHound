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

import { faFolderOpen, faPlay, faQuestion } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Button, Collapse } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import '@neo4j-cypher/codemirror/css/cypher-codemirror.css';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import {
    ActiveDirectoryNodeKind,
    ActiveDirectoryRelationshipKind,
    AzureNodeKind,
    AzureRelationshipKind,
    ActiveDirectoryKindProperties,
    AzureKindProperties,
} from 'bh-shared-ui';
import { useCallback, useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { startCypherQuery } from 'src/ducks/explore/actions';
import { setCypherQueryTerm, startCypherSearch } from 'src/ducks/searchbar/actions';
import { CommonKindProperties } from 'src/graphSchema';
import { AppState } from 'src/store';
import CommonSearches from './CommonSearches';

const useStyles = makeStyles((theme) => ({
    button: {
        minWidth: '35px',
        height: '35px',
        borderRadius: theme.shape.borderRadius,
        borderColor: 'rgba(0,0,0,0.23)',
        color: 'black',
    },
    textarea: {
        resize: 'vertical',
        fontSize: '.8rem',
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

const CypherInput = () => {
    const classes = useStyles();
    const dispatch = useDispatch();
    const [toggle, setToggle] = useState(false);

    const [cypherQuery, setCypherQuery] = useState('');

    const syncCypherQueryWithAppState = useCallback(
        () => dispatch(setCypherQueryTerm(cypherQuery)),
        [cypherQuery, dispatch]
    );

    const cypherAppState = useSelector((state: AppState) => state.search.cypher);

    useEffect(() => {
        if (cypherAppState.searchTerm) {
            setCypherQuery(cypherAppState.searchTerm);
        }
    }, [cypherAppState]);

    useEffect(() => {
        syncCypherQueryWithAppState();
    }, [syncCypherQueryWithAppState]);
    const [showCommonQueries, setShowCommonQueries] = useState(false);

    const onClickCommonSearchesListItem = (query: string) => {
        dispatch(setCypherQueryTerm(query));
        dispatch(startCypherQuery(query));
    };

    const handleCypherSearch = (cypherQuery?: string) => {
        if (cypherQuery) {
            // Easter Egg Trigger
            if (cypherQuery.toLowerCase().includes('match (n) return n limit 5')) {
                setToggle(true);
            } else {
                setToggle(false);
            }
            dispatch(startCypherSearch(cypherQuery));
        }
    };

    return (
        <>
            <Box display={'flex'} gap={1}>
                <Button
                    className={classes.button}
                    sx={{ padding: 0 }}
                    onClick={() => {
                        setShowCommonQueries((v) => !v);
                    }}
                    variant='outlined'>
                    <FontAwesomeIcon icon={faFolderOpen} />
                </Button>

                <div
                    onClick={() => {
                        const input = document.querySelector('.cm-content') as HTMLElement;
                        if (input) {
                            input.focus();
                        }
                    }}
                    style={{ flex: 1 }}
                    role='textbox'>
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
                                handleCypherSearch(cypherQuery);
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
                <a
                    href='https://support.bloodhoundenterprise.io/hc/en-us/articles/16721164740251'
                    target='_blank'
                    rel='noreferrer'>
                    <Button variant='outlined' className={classes.button} sx={{ padding: 0 }}>
                        <FontAwesomeIcon icon={faQuestion} />
                    </Button>
                </a>

                <Button className={classes.button} onClick={() => handleCypherSearch(cypherQuery)} variant='outlined'>
                    <Box mr={1}>
                        <FontAwesomeIcon icon={faPlay} />
                    </Box>
                    Search
                </Button>
            </Box>
            <Collapse in={showCommonQueries}>
                <CommonSearches onClickListItem={onClickCommonSearchesListItem} />
            </Collapse>

            {
                /*What is graphed will never die*/ toggle && (
                    <div style={{ display: 'block', position: 'relative', bottom: 0, left: 0 }}>
                        <img
                            src={'img/logo-animated.gif'}
                            alt='What is graphed will never die'
                            style={{ width: '200px' }}
                        />
                    </div>
                )
            }
        </>
    );
};

export default CypherInput;
