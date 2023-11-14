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
import { Typography } from '@mui/material';
import { useQuery } from 'react-query';
import { apiClient } from '../../..';
import { EdgeInfoProps } from '..';

const Details: FC<EdgeInfoProps> = ({ sourceDBId, targetDBId, edgeName }) => {
    console.log({ sourceDBId, targetDBId, edgeName });

    const { data, isLoading, isError } = useQuery(['edge', edgeName, sourceDBId, targetDBId], () =>
        apiClient.getEdgeDetails(sourceDBId!, targetDBId!, edgeName!)
    );

    console.log(data, isLoading, isError);

    return (
        <>
            <Typography variant='body2'>
                The relationship represents the effective outcome of the configuration and relationships between several
                different objects. All objects involved in the creation of this relationship are listed here:
            </Typography>
            <ul>
                {Object.entries(data?.data.data.nodes || {}).map(([key, value], index) => (
                    <li key={key}>{value.label}</li>
                ))}
            </ul>
        </>
    );
};

export default Details;
