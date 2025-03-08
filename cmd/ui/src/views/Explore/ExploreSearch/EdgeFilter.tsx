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
import { faFilter } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import makeStyles from '@mui/styles/makeStyles';
import { useState } from 'react';
import EdgeFilteringDialog from './EdgeFilteringDialog';
import { usePathfindingFilterSwitch } from './switches';

const useStyles = makeStyles((theme) => ({
    pathfindingButton: {
        height: '25px',
        width: '25px',
        minWidth: '25px',
        borderRadius: theme.shape.borderRadius,
        borderColor: 'rgba(0,0,0,0.23)',
        color: theme.palette.common.white,
        padding: 0,
    },
}));

const EdgeFilter = () => {
    const classes = useStyles();

    const [isOpenDialog, setIsOpenDialog] = useState(false);

    const { selectedFilters, initialize, handleApplyFilters, handleUpdateFilters, handleCancelFilters } =
        usePathfindingFilterSwitch();

    return (
        <>
            <Button
                className={classes.pathfindingButton}
                onClick={() => {
                    setIsOpenDialog(true);
                    // what is the initial state of edge filters?  save it
                    initialize();
                }}>
                <FontAwesomeIcon icon={faFilter} />
            </Button>
            <EdgeFilteringDialog
                isOpen={isOpenDialog}
                selectedFilters={selectedFilters}
                handleApply={() => {
                    setIsOpenDialog(false);
                    handleApplyFilters();
                }}
                handleUpdate={handleUpdateFilters}
                handleCancel={() => {
                    setIsOpenDialog(false);
                    handleCancelFilters();
                }}
            />
        </>
    );
};

export default EdgeFilter;
