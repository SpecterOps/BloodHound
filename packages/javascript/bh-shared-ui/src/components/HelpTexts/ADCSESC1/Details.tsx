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

import { FC } from 'react';
import { Alert, Box, List, ListItem, Skeleton, Tooltip, Typography } from '@mui/material';
import { NodeIcon, apiClient } from '../../..';
import { EdgeInfoProps } from '..';
import { FixedSizeList } from 'react-window';
import { useQuery } from 'react-query';

const Details: FC<EdgeInfoProps> = ({ sourceDBId, targetDBId, edgeName }) => {
    const { data, isLoading, isError } = useQuery(['edgeDetails', sourceDBId, targetDBId, edgeName], ({ signal }) =>
        apiClient.getEdgeDetails(sourceDBId!, targetDBId!, edgeName!).then((result) => result.data)
    );

    const InnerElement = ({ style, ...rest }: any) => (
        <List component='ul' disablePadding style={{ ...style, overflowX: 'hidden' }} {...rest} />
    );

    const Row = ({ data, index, style }: any) => {
        const items = data;
        const item = items[index];
        const itemClass = index % 2 ? 'odd-item' : 'even-item';

        if (item === undefined) {
            return (
                <ListItem
                    className={itemClass}
                    style={{ ...style, whiteSpace: 'nowrap', padding: '0 8px' }}
                    data-testid='entity-row'>
                    <Skeleton variant='text' width='100%' />
                </ListItem>
            );
        }

        const normalizedItem = {
            id: item.objectID || item.props?.objectid || '',
            name: item.name || item.objectID || item.props?.name || item.props?.objectid || 'Unknown',
            type: item.label || item.kind || '',
        };

        return (
            <ListItem
                button
                className={itemClass}
                style={{
                    ...style,
                    padding: '0 8px',
                }}
                data-testid='entity-row'>
                <NodeIcon nodeType={normalizedItem.type} />
                <Tooltip title={normalizedItem.name}>
                    <div style={{ minWidth: '0', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {normalizedItem.name}
                    </div>
                </Tooltip>
            </ListItem>
        );
    };

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
                    <Alert severity='error'>Couldn't load edge details</Alert>
                ) : (
                    <FixedSizeList
                        height={Math.min(Object.keys(data?.data.nodes || {}).length, 16) * 32}
                        itemCount={Object.keys(data?.data.nodes || {}).length}
                        itemData={Object.entries(data?.data.nodes || {}).map(([key, value]) => ({
                            objectID: value.objectId,
                            name: value.label,
                            kind: value.kind,
                        }))}
                        itemSize={32}
                        innerElementType={InnerElement}
                        width={'100%'}
                        initialScrollOffset={0}
                        style={{ borderRadius: 4 }}>
                        {Row}
                    </FixedSizeList>
                )}
            </Box>
        </>
    );
};

export default Details;
