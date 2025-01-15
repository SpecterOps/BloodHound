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

import { Box, Skeleton, Typography } from '@mui/material';
import { apiClient } from '../utils';
import React from 'react';
import { useQuery } from 'react-query';

const ApiVersion: React.FC = () => {
    const getVersionQuery = useQuery(['Version'], ({ signal }) =>
        apiClient.version({ signal }).then((res) => res.data.data)
    );

    const version = getVersionQuery;

    if (getVersionQuery.isLoading)
        return (
            <Box paddingX={2}>
                <Skeleton variant='text' />
            </Box>
        );

    if (getVersionQuery.isError)
        return <Typography variant='body2'>API Version: Unknown, please refresh the page or report a bug.</Typography>;

    return <Typography variant='body2'>API Version: {version.data.server_version}</Typography>;
};

export default ApiVersion;
