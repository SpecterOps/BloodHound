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
import { useQuery } from 'react-query';
import { apiClient } from '../../../utils';
import VirtualizedNodeList, { VirtualizedNodeListItem } from '../../VirtualizedNodeList';
import { EdgeInfoProps } from '../index';
import {searchbarActions} from "../../../store";
import { useDispatch } from 'react-redux';

const RelayTargets: FC<EdgeInfoProps> = ({ sourceDBId, targetDBId, edgeName }) => {
    const { data, isLoading, isError } = useQuery(['relayTargets', sourceDBId, targetDBId, edgeName], () =>
        apiClient.getRelayTargets(sourceDBId!, targetDBId!, edgeName!).then((result) => result.data)
    );

    const dispatch = useDispatch();

    const handleOnClick = (item: any) => {
        let node = nodesArray[item]
        dispatch(
            searchbarActions.sourceNodeSelected({
                objectid: node.objectId,
                type: node.kind,
                name: node.name,
            })
        );

        dispatch(searchbarActions.tabChanged("primary"))
    };

    const nodesArray: VirtualizedNodeListItem[] = Object.values(data?.data.nodes || {}).map((node) => ({
        name: node.label,
        objectId: node.objectId,
        kind: node.kind,
        onClick: handleOnClick
    }));

    return (
        <>
            <Typography variant='body2'>The nodes in this list are valid relay targets for this attack</Typography>
            <Box py={1}>
                {isLoading ? (
                    <Skeleton variant='rounded' />
                ) : isError ? (
                    <Alert severity='error'>Couldn't load relay targets</Alert>
                ) : (
                    <VirtualizedNodeList nodes={nodesArray} />
                )}
            </Box>
        </>
    );
};

export default RelayTargets;
