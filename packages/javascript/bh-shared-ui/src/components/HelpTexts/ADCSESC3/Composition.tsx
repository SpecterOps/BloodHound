// Copyright 2024 Specter Ops, Inc.
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

import { FC } from 'react';
import { Alert, Box, Skeleton, Typography } from '@mui/material';
import { apiClient } from '../../../utils/api';
import { EdgeInfoProps } from '..';
import { useQuery } from 'react-query';
import VirtualizedNodeList, { VirtualizedNodeListItem } from '../../VirtualizedNodeList';

const Composition: FC<EdgeInfoProps> = ({ sourceDBId, targetDBId, edgeName }) => {
    const { data, isLoading, isError } = useQuery(['edgeComposition', sourceDBId, targetDBId, edgeName], ({ signal }) =>
        apiClient.getEdgeComposition(sourceDBId!, targetDBId!, edgeName!).then((result) => result.data)
    );

    const nodesArray: VirtualizedNodeListItem[] = Object.values(data?.data.nodes || {}).map((node) => ({
        name: node.label,
        objectId: node.objectId,
        kind: node.kind,
    }));

    return (
        <>
            <Typography variant='body2'>
                The relationship represents the effective outcome of the configuration and relationships between several
                different objects. All objects involved in the creation of this relationship are listed here:
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
