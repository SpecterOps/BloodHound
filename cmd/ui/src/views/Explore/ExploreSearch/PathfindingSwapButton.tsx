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

import { faExchangeAlt } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { useCallback } from 'react';
import { destinationNodeSelected, sourceNodeSelected } from 'src/ducks/searchbar/actions';
import { useAppDispatch, useAppSelector } from 'src/store';

const useStyles = makeStyles((theme) => ({
    swapButton: {
        height: '25px',
        width: '25px',
        minWidth: '25px',
        borderRadius: theme.shape.borderRadius,
        borderColor: 'rgba(0,0,0,0.23)',
        color: 'black',
        padding: 0,
    },
}));

const PathfindingSwapButton = () => {
    const classes = useStyles();
    const dispatch = useAppDispatch();

    const { primary, secondary } = useAppSelector((state) => state.search);

    const swapPathfindingInputs = useCallback(() => {
        const newSourceNode = secondary.value;
        const newDestinationNode = primary.value;

        dispatch(sourceNodeSelected(newSourceNode));
        dispatch(destinationNodeSelected(newDestinationNode));
    }, [primary, secondary, dispatch]);

    return (
        <Button
            className={classes.swapButton}
            variant='outlined'
            disabled={!primary?.value || !secondary?.value}
            onClick={swapPathfindingInputs}>
            <FontAwesomeIcon icon={faExchangeAlt} className='fa-rotate-90' />
        </Button>
    );
};

export default PathfindingSwapButton;
