// Copyright 2025 Specter Ops, Inc.
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

import { Alert, Box, Skeleton, Typography } from '@mui/material';
import { FC } from 'react';
import { EdgeInfoType, useEdgeInfoItems } from '../../../hooks/useEdgeInfoItems';
import VirtualizedNodeList from '../../VirtualizedNodeList';
import { EdgeInfoProps } from '../index';

const RelaySources: FC<EdgeInfoProps> = ({ sourceDBId, targetDBId, edgeName }) => {
    const { isLoading, isError, nodesArray } = useEdgeInfoItems({
        sourceDBId,
        targetDBId,
        edgeName,
        type: EdgeInfoType['relayTargets'],
    });

    return (
        <>
            <Typography variant='body2'>The nodes in this list are valid relay sources for this attack</Typography>
            <Box py={1}>
                {isLoading ? (
                    <Skeleton variant='rounded' />
                ) : isError ? (
                    <Alert severity='error'>Couldn't load coercion targets</Alert>
                ) : (
                    <VirtualizedNodeList nodes={nodesArray} />
                )}
            </Box>
        </>
    );
};

export default RelaySources;
