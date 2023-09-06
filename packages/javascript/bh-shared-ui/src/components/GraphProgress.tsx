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

import { LinearProgress } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import React from 'react';

const useStyles = makeStyles({
    container: {
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
    },
    progressRoot: {
        backgroundColor: '#6798b9',

        '& .MuiLinearProgress-barColorPrimary': {
            backgroundColor: '#406f8e',
        },
    },
});

const GraphProgress: React.FC<{
    loading: boolean;
}> = ({ loading }) => {
    const styles = useStyles();

    if (!loading) return null;

    return (
        <div className={styles.container}>
            <LinearProgress color='primary' className={styles.progressRoot} />
        </div>
    );
};

export default GraphProgress;
