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
import { EntityInfoDataTable, EntityInfoPanel } from '../../../components';
import { useAssetGroupTagInfo, useMemberInfo, usePZPathParams, useRuleInfo } from '../../../hooks';
import { EntityKinds } from '../../../utils';
import DynamicDetailsV2 from './DynamicDetailsV2';
import EntityRulesInformation from './EntityRulesInformation';

type SelectedDetailsV2Props = {
    currentTab: string;
};

export const SelectedDetailsV2: FC<SelectedDetailsV2Props> = ({ currentTab }) => {
    const { ruleId, memberId, tagId } = usePZPathParams();

    const tagQuery = useAssetGroupTagInfo(tagId);

    const ruleQuery = useRuleInfo(tagId, ruleId);

    const memberQuery = useMemberInfo(tagId, memberId);

    if (memberQuery.data && currentTab === '3') {
        const selectedNode = {
            id: memberQuery.data.object_id,
            name: memberQuery.data.name,
            type: memberQuery.data.primary_kind as EntityKinds,
        };
        return (
            <div className='h-full'>
                <EntityInfoPanel
                    DataTable={EntityInfoDataTable}
                    selectedNode={selectedNode}
                    additionalTables={[
                        {
                            sectionProps: { id: memberQuery.data.object_id, label: 'Rules' },
                            TableComponent: EntityRulesInformation,
                        },
                    ]}
                />
            </div>
        );
    } else if (ruleId !== undefined && currentTab === '2') {
        return <DynamicDetailsV2 queryResult={ruleQuery} />;
    } else if (tagId !== undefined && currentTab === '1') {
        return <DynamicDetailsV2 queryResult={tagQuery} />;
    }

    return null;
};
