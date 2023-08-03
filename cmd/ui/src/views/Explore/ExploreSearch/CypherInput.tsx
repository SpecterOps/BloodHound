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
import { Box, Button, Collapse, TextField } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { useCallback, useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { setCypherQueryTerm, startCypherSearch } from 'src/ducks/searchbar/actions';
import { AppState } from 'src/store';
import CommonSearches from './CommonSearches';
import { startCypherQuery } from 'src/ducks/explore/actions';

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
}));

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

                <Box display='flex' flexGrow={1} flexDirection={'column'} gap={1} alignItems={'end'}>
                    <TextField
                        placeholder='Cypher Search'
                        multiline
                        rows={3}
                        fullWidth
                        value={cypherQuery}
                        onChange={(e) => {
                            setCypherQuery(e.target.value);
                        }}
                        onKeyDown={(e) => {
                            // if enter and shift key is pressed, execute cypher search
                            if (e.key === 'Enter' && e.shiftKey) {
                                e.preventDefault();
                                handleCypherSearch(cypherQuery);
                            }
                        }}
                        inputProps={{
                            style: { fontFamily: 'monospace' },
                            className: classes.textarea,
                        }}
                    />
                    <Box display={'flex'} gap={1}>
                        <a
                            href='https://support.bloodhoundenterprise.io/hc/en-us/articles/16721164740251'
                            target='_blank'
                            rel='noreferrer'>
                            <Button variant='outlined' className={classes.button} sx={{ padding: 0 }}>
                                <FontAwesomeIcon icon={faQuestion} />
                            </Button>
                        </a>

                        <Button
                            className={classes.button}
                            onClick={() => handleCypherSearch(cypherQuery)}
                            variant='outlined'>
                            <Box mr={1}>
                                <FontAwesomeIcon icon={faPlay} />
                            </Box>
                            Search
                        </Button>
                    </Box>
                </Box>
            </Box>

            <Collapse in={showCommonQueries}>
                <CommonSearches onClickListItem={onClickCommonSearchesListItem} />
            </Collapse>

            {
                /*What is graphed will never die*/ toggle && (
                    <div style={{ display: 'block', position: 'relative', bottom: 0, left: 0 }}>
                        <img
                            src={`${import.meta.env.BASE_URL}/img/logo-animated.gif`}
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
