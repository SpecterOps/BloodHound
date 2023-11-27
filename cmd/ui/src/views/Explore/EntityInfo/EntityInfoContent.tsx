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

import { Box } from '@mui/material';
import { allSections, entityInformationEndpoints, EntityKinds } from 'bh-shared-ui';
import React from 'react';
import { useQuery } from 'react-query';
import EntityInfoDataTableList from './EntityInfoDataTableList';
import EntityObjectInformation from './EntityObjectInformation';

export interface EntityInfoContentProps {
    id: string;
    nodeType: EntityKinds;
}

const EntityInfoContent: React.FC<EntityInfoContentProps> = ({ id, nodeType }) => {
    const { data, isLoading, isError } = useQuery(
        ['entity', nodeType, id],
        ({ signal }) => entityInformationEndpoints[nodeType]?.(id, { signal }).then((res) => res.data.data),
        { refetchOnWindowFocus: false, retry: false }
    );

    const sections = allSections[nodeType]?.(id);

    return (
        <Box>
            {isLoading ? (
                <div>Loading...</div>
            ) : isError || data === undefined || sections === undefined ? (
                <div>Unable to load node information.</div>
            ) : (
                <EntityObjectInformation {...data} />
            )}
            <EntityInfoDataTableList tables={sections || []} />
        </Box>
    );
};

export default EntityInfoContent;
