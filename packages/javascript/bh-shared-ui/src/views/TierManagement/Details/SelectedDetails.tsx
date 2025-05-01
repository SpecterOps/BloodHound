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

import { SeedTypeCypher } from 'js-client-library';
import { FC } from 'react';
import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { apiClient } from '../../../utils';
import { Cypher } from '../Cypher';
import DynamicDetails from './DynamicDetails';
import WrappedEntityInfoPanel from './EntityInfo/EntityInfoPanel';
import ObjectCountPanel from './ObjectCountPanel';
import { getSelectorSeedType } from './utils';

export const SelectedDetails: FC = () => {
    const { selectorId, memberId, tagId } = useParams();

    const tagQuery = useQuery({
        queryKey: ['tier-management', 'tag', tagId],
        queryFn: async () => {
            if (!tagId) return undefined;
            return apiClient.getAssetGroupTag(tagId).then((res) => {
                return res.data.data['tag'];
            });
        },
        enabled: tagId !== undefined,
    });

    const selectorQuery = useQuery({
        queryKey: ['tier-management', 'tags', tagId, 'selectors', selectorId],
        queryFn: async () => {
            if (!tagId || !selectorId) return undefined;
            return apiClient.getAssetGroupTagSelector(tagId, selectorId).then((res) => {
                return res.data.data['selector'];
            });
        },
        enabled: tagId !== undefined && selectorId !== undefined,
    });

    const memberQuery = useQuery({
        queryKey: ['tier-management', 'tags', tagId, 'member', memberId],
        queryFn: async () => {
            if (!tagId || !memberId) return undefined;
            return apiClient.getAssetGroupTagMemberInfo(tagId, memberId).then((res) => {
                return res.data.data['member'];
            });
        },
        enabled: tagId !== undefined && memberId !== undefined,
    });

    if (tagId !== undefined && memberId !== undefined && memberQuery.data !== undefined) {
        return <WrappedEntityInfoPanel selectedNode={memberQuery.data} />;
    }

    if (selectorId !== undefined && selectorQuery.data !== undefined) {
        return (
            <div className='max-h-full flex flex-col gap-8'>
                <DynamicDetails data={selectorQuery.data} />
                {getSelectorSeedType(selectorQuery.data) === SeedTypeCypher && (
                    <Cypher preview initialInput={selectorQuery.data.seeds[0].value} />
                )}
            </div>
        );
    }

    if (tagId !== undefined && tagQuery.data !== undefined) {
        return (
            <div className='max-h-full flex flex-col gap-8'>
                <DynamicDetails data={tagQuery.data} />
                <ObjectCountPanel tagId={tagId} />
            </div>
        );
    }

    return null;
};
