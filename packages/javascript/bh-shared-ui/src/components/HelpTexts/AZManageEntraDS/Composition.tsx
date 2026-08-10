// Copyright 2026 Specter Ops, Inc.
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

import { Alert, Box, Skeleton } from '@mui/material';
import { Typography } from 'doodle-ui';
import { FC } from 'react';
import { EdgeInfoProps } from '..';
import { EdgeInfoItems, useEdgeInfoItems } from '../../../hooks/useExploreGraph/useEdgeInfoItems';
import VirtualizedNodeList from '../../VirtualizedNodeList';

const Composition: FC<EdgeInfoProps> = ({ sourceDBId, targetDBId, edgeName }) => {
    const { isLoading, isError, nodesArray } = useEdgeInfoItems({
        sourceDBId,
        targetDBId,
        edgeName,
        type: EdgeInfoItems['composition'],
    });

    return (
        <>
            <Typography variant='body2'>
                The relationship combines the source principal's effective Contributor or Domain Services Contributor
                path to this managed domain with its Application Administrator and Groups Administrator role paths. All
                three permission components must apply to the same source principal. Tenant-to-role containment is
                neither required nor included in this composition.
            </Typography>
            <Box py={1}>
                {isLoading ? (
                    <Skeleton variant='rounded' />
                ) : isError ? (
                    <Alert severity='error'>Couldn't load edge composition</Alert>
                ) : (
                    <VirtualizedNodeList nodes={nodesArray} />
                )}
            </Box>
        </>
    );
};

export default Composition;
