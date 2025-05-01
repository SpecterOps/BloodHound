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
import React from 'react';
import { useParams } from 'react-router-dom';
import EntityObjectInformation from './EntityObjectInformation';
import SelectorList from './SelectorList';

export interface EntityInfoContentProps {
    id: number;
    nodeType: string;
    properties: Record<string, any>;
}

const EntityInfoContent: React.FC<EntityInfoContentProps> = (props) => {
    const { tagId, memberId } = useParams();

    return (
        <Box>
            <EntityObjectInformation {...props} />
            {tagId !== undefined && memberId !== undefined && <SelectorList tagId={tagId} memberId={memberId} />}
        </Box>
    );
};

export default EntityInfoContent;
