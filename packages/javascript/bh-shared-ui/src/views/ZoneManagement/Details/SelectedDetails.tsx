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

import { FC } from 'react';
import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { apiClient } from '../../../utils';
import DynamicDetails from './DynamicDetails';
import WrappedEntityInfoPanel from './EntityInfo/EntityInfoPanel';

export const SelectedDetails: FC = () => {
    const { tierId, labelId, selectorId, memberId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;

    const tagQuery = useQuery({
        queryKey: ['zone-management', 'tag', tagId],
        queryFn: async () => {
            if (!tagId) return undefined;
            return apiClient.getAssetGroupTag(tagId).then((res) => {
                return res.data.data['tag'];
            });
        },
        enabled: tagId !== undefined,
    });

    const selectorQuery = useQuery({
        queryKey: ['zone-management', 'tags', tagId, 'selectors', selectorId],
        queryFn: async () => {
            if (!tagId || !selectorId) return undefined;
            return apiClient.getAssetGroupTagSelector(tagId, selectorId).then((res) => {
                return res.data.data['selector'];
            });
        },
        enabled: tagId !== undefined && selectorId !== undefined,
    });

    const memberQuery = useQuery({
        queryKey: ['zone-management', 'tags', tagId, 'member', memberId],
        queryFn: async () => {
            if (!tagId || !memberId) return undefined;
            return apiClient.getAssetGroupTagMemberInfo(tagId, memberId).then((res) => {
                return res.data.data['member'];
            });
        },
        enabled: tagId !== undefined && memberId !== undefined,
    });

    if (memberId !== undefined) {
        return <WrappedEntityInfoPanel selectedNode={memberQuery.data ?? null} />;
    } else if (selectorId !== undefined) {
        return <DynamicDetails queryResult={selectorQuery} />;
    } else if (tagId !== undefined) {
        return <DynamicDetails queryResult={tagQuery} />;
    }

    return null;
};
