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
import { AssetGroupTag, AssetGroupTagSelector } from 'js-client-library';
import { FC } from 'react';

type DynamicDetailsProps = {
    data: AssetGroupTagSelector | AssetGroupTag | undefined;
    isCypher?: boolean;
};

const isSelector = (data: any): data is AssetGroupTagSelector => {
    return 'seeds' in data;
};

const isLabel = (data: any): data is AssetGroupTag => {
    return 'asset_group_tier_id' in data;
};

const DynamicDetails: FC<DynamicDetailsProps> = ({ data, isCypher }) => {
    if (!data) {
        return null;
    }
    const lastUpdated = new Date(data.updated_at).toLocaleDateString();

    return (
        <Card className='h-[280px] mb-[24px] px-6 pt-6 select-none overflow-y-auto'>
            <div className='text-xl font-bold'>{data ? data.name : 'Nothing Data'}</div>
            <div className='flex flex-wrap gap-x-2'>
                {isLabel(data) && data.position !== null && (
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
            {isLabel(data) && (
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
