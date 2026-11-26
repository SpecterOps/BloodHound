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

import { Card } from 'doodle-ui';
import { FC } from 'react';
import { EntityInfoDataTable, EntityInfoPanel } from '../../../../components';
import { useAssetGroupTagInfo, useGetNodeById, useMemberInfo, useRuleInfo } from '../../../../hooks';
import { DetailsTabOption, ObjectTabValue, RuleTabValue, TagTabValue } from '../../utils';
import DynamicDetails from '../DynamicDetails';
import EntityRulesInformation from '../EntityRulesInformation';
import { OpenObjectInExploreLink } from '../OpenObjectInExploreLink';

type SelectedDetailsTabContent = {
    currentDetailsTab: DetailsTabOption;
    tagId: string;
    memberId?: string;
    ruleId?: string;
};

export const SelectedDetailsTabContent: FC<SelectedDetailsTabContent> = ({
    currentDetailsTab,
    ruleId,
    memberId,
    tagId,
}) => {
    const tagQuery = useAssetGroupTagInfo(tagId);

    const ruleQuery = useRuleInfo(tagId, ruleId);

    const memberQuery = useMemberInfo(tagId, memberId);
    const { data: selectedNode, isPreviousData } = useGetNodeById(memberQuery.data?.id);

    if (memberQuery.data && currentDetailsTab === ObjectTabValue) {
        const { object_id: objectId, id: nodeId } = memberQuery.data;
        return (
            <div className='inline-flex h-full flex-col gap-2'>
                <Card className='shrink-0 px-4 py-2'>
                    <OpenObjectInExploreLink nodeId={nodeId} />
                </Card>
                <EntityInfoPanel
                    className='min-h-0'
                    panelClassName='bg-neutral-1 border border-solid border-neutral-light-4 dark:border-neutral-900'
                    selectedNode={isPreviousData ? undefined : selectedNode}
                    showPlaceholderMessage={true}
                    DataTable={EntityInfoDataTable}
                    additionalTables={[
                        {
                            sectionProps: { id: objectId, label: 'Rules' },
                            TableComponent: EntityRulesInformation,
                        },
                    ]}
                />
            </div>
        );
    } else if (ruleId !== undefined && currentDetailsTab === RuleTabValue) {
        return <DynamicDetails queryResult={ruleQuery} />;
    } else if (tagId !== undefined && currentDetailsTab === TagTabValue) {
        return <DynamicDetails queryResult={tagQuery} />;
    }

    return null;
};
