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
import { ActiveDirectoryKindProperties } from '../../../graphSchema';
import { EdgeInfoItems, useEdgeInfoItems } from '../../../hooks/useExploreGraph/useEdgeInfoItems';
import VirtualizedNodeList, { NormalizedNodeItem } from '../../VirtualizedNodeList';

type ACLInheritanceListProps = {
    sourceDBId: number;
    targetDBId: number;
    edgeName: string;
    inheritanceHash: string;
};

const ACLInheritance: FC<ACLInheritanceListProps> = ({ sourceDBId, targetDBId, edgeName, inheritanceHash }) => {
    const { isLoading, isError, nodesArray } = useEdgeInfoItems(
        {
            sourceDBId,
            targetDBId,
            edgeName,
            type: EdgeInfoItems['aclInheritance'],
        },
        { withProperties: true }
    );

    const checkNodeForHash = (node: NormalizedNodeItem) => {
        const hashes = node.properties?.[ActiveDirectoryKindProperties.InheritanceHashes];
        return !!(Array.isArray(hashes) && hashes.includes(inheritanceHash));
    };

    const nodesInheritedFrom = inheritanceHash.length > 0 ? nodesArray.filter(checkNodeForHash) : [];

    return (
        <>
            <Typography variant='body2'>Maybe some text about what ACL Inheritance means:</Typography>
            <Box py={1}>
                {isLoading ? (
                    <Skeleton variant='rounded' />
                ) : isError ? (
                    <Alert severity='error'>Couldn't load ACL inheritance</Alert>
                ) : (
                    <VirtualizedNodeList nodes={nodesInheritedFrom} />
                )}
            </Box>
        </>
    );
};

export default ACLInheritance;
