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

export const INHERITANCE_DROPDOWN_DESCRIPTION =
    'An ACE granting access permissions to a Domain, OU, or Container can be inherited by entities contained within them. This panel lists the source object(s) for the inherited ACE.';

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

    // Filter our source node(s) from the result set by checking if the node's inheritance hash list includes the target edge's hash
    const checkNodeForHash = (node: NormalizedNodeItem) => {
        const hashes = node.properties?.[ActiveDirectoryKindProperties.InheritanceHashes];
        return !!(Array.isArray(hashes) && hashes.includes(inheritanceHash));
    };

    const nodesInheritedFrom = inheritanceHash.length > 0 ? nodesArray.filter(checkNodeForHash) : [];

    const getSourceObjectContent = () => {
        if (isLoading) {
            return <Skeleton variant='rounded' />;
        }

        if (isError) {
            return <Alert severity='error'>Couldn't load ACL inheritance sources</Alert>;
        }

        if (nodesInheritedFrom.length === 0) {
            return <Alert severity='warning'>No valid ACL inheritance sources found</Alert>;
        }

        return <VirtualizedNodeList nodes={nodesInheritedFrom} />;
    };

    return (
        <>
            <Typography variant='body2'>{INHERITANCE_DROPDOWN_DESCRIPTION}</Typography>
            <Box py={1}>{getSourceObjectContent()}</Box>
        </>
    );
};

export default ACLInheritance;
