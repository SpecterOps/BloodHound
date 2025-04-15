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

import { Card } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTag, AssetGroupTagSelector, AssetGroupTagSelectorMember } from 'js-client-library';
import { DateTime } from 'luxon';
import { FC } from 'react';
import { useParams } from 'react-router-dom';
import { LuxonFormat } from '../../../utils';
import { Cypher } from '../Cypher';
import WrappedEntityInfoPanel from './EntityInfo/EntityInfoPanel';
import ObjectCountPanel from './ObjectCountPanel';
import { isMember, isSelector, isTag } from './utils';

type SelectedDetailsProps = {
    selectedTag: AssetGroupTag | undefined;
    selectedSelector: AssetGroupTagSelector | undefined;
    selectedMember: AssetGroupTagSelectorMember | null;
    cypher?: boolean;
};

export const SelectedDetails: FC<SelectedDetailsProps> = ({
    selectedTag,
    selectedSelector,
    selectedMember,
    cypher,
}) => {
    const { selectorId, memberId, tagId } = useParams();

    if (memberId !== undefined) {
        if (isMember(selectedMember)) {
            return <WrappedEntityInfoPanel selectedNode={selectedMember} />;
        }
    }

    if (selectorId !== undefined) {
        if (cypher)
            return (
                <>
                    <DynamicDetails data={selectedSelector} isCypher={cypher} />
                    <Cypher />
                </>
            );
        else return <DynamicDetails data={selectedSelector} />;
    }

    if (tagId !== undefined) {
        return (
            <>
                <DynamicDetails data={selectedTag} />
                <ObjectCountPanel tagId={tagId} />
            </>
        );
    }

    return null;
};

type DynamicDetailsProps = {
    data: AssetGroupTagSelector | AssetGroupTag | undefined;
    isCypher?: boolean;
};

const DynamicDetails: FC<DynamicDetailsProps> = ({ data, isCypher }) => {
    if (!data) {
        return null;
    }

    const lastUpdated = DateTime.fromISO(data.updated_at).toFormat(LuxonFormat.ISO_8601_SLASHES);

    return (
        <Card className='h-64 mb-8 px-6 pt-6 max-w-[32rem] select-none overflow-y-auto'>
            <div className='text-xl font-bold'>{data ? data.name : 'Nothing Data'}</div>
            <div className='flex flex-wrap gap-x-2'>
                {isTag(data) && data.position !== null && (
                    <>
                        <p className='font-bold'>Tier:</p> <p>{data.position}</p>
                    </>
                )}
            </div>
            <div className='flex flex-wrap gap-x-2'>
                <p className='font-bold'>Description:</p> <p>{data.description}</p>
            </div>
            <div className='flex flex-wrap gap-x-2'>
                <p className='font-bold'>Created by:</p> <p>{data.created_by}</p>
            </div>
            <div className='flex flex-wrap gap-x-2'>
                <p className='font-bold'>Last Updated:</p> <p>{lastUpdated}</p>
            </div>
            {isTag(data) && (
                <div className='flex flex-wrap gap-x-2'>
                    <p className='font-bold'>Certification Enabled:</p> <p>{data.requireCertify}</p>
                </div>
            )}
            {isSelector(data) && (
                <>
                    <div className='flex flex-wrap gap-x-2'>
                        <p className='font-bold'>Type:</p> <p>{isCypher ? 'Cypher' : 'Object'}</p>
                    </div>
                    <div className='flex flex-wrap gap-x-2'>
                        <p className='font-bold'>Automatic Certification:</p>{' '}
                        <p>{data.auto_certify ? 'Enabled' : 'Disabled'}</p>
                    </div>
                    <div className='flex flex-wrap gap-x-2'>
                        <p className='font-bold'>Selector Enabled:</p>{' '}
                        <p>{data.disabled_at ? 'Enabled' : 'Disabled'}</p>
                    </div>
                </>
            )}
        </Card>
    );
};

export default DynamicDetails;
