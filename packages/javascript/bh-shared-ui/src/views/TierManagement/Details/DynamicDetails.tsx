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
import { AssetGroupTag, AssetGroupTagSelector, SeedTypesMap } from 'js-client-library';
import { DateTime } from 'luxon';
import { FC } from 'react';
import { LuxonFormat } from '../../../utils';
import { getSelectorSeedType, isSelector, isTag } from './utils';

type DynamicDetailsProps = {
    data: AssetGroupTag | AssetGroupTagSelector;
};

const DetailField: FC<{ label: string; value: string }> = ({ label, value }) => {
    return (
        <div className='flex flex-wrap gap-x-2'>
            <span className='font-bold'>{label}:</span>
            <span className='truncate text-ellipsis' title={value}>
                {value}
            </span>
        </div>
    );
};

const TagDetails: FC<{ data: AssetGroupTag }> = ({ data }) => {
    const lastUpdated = DateTime.fromISO(data.updated_at).toFormat(LuxonFormat.YEAR_MONTH_DAY_SLASHES);

    return (
        <Card className='px-6 py-6 max-w-[32rem]'>
            <div className='text-xl font-bold truncate'>{data.name}</div>
            {data.position !== null && (
                <div className='mt-4'>
                    <DetailField label='Tier' value={data.position.toString()} />
                </div>
            )}
            <div className='mt-4'>
                <DetailField label='Description' value={data.description} />
            </div>
            <div className='mt-4'>
                <DetailField label='Created by' value={data.created_by} />
                <DetailField label='Last Updated' value={lastUpdated} />
            </div>
            <div className='mt-4'>
                <DetailField label='Certification' value={data.requireCertify ? 'Required' : 'Not Required'} />
            </div>
        </Card>
    );
};

const SelectorDetails: FC<{ data: AssetGroupTagSelector }> = ({ data }) => {
    const lastUpdated = DateTime.fromISO(data.updated_at).toFormat(LuxonFormat.YEAR_MONTH_DAY_SLASHES);
    const seedType = getSelectorSeedType(data);

    return (
        <Card className='px-6 py-6 max-w-[32rem]'>
            <div className='text-xl font-bold truncate'>{data.name}</div>
            <div className='mt-4'>
                <DetailField label='Description' value={data.description} />
            </div>
            <div className='mt-4'>
                <DetailField label='Created by' value={data.created_by} />
                <DetailField label='Last Updated' value={lastUpdated} />
                <DetailField label='Type' value={SeedTypesMap[seedType]} />
            </div>
            <div className='mt-4'>
                <DetailField label='Automatic Certification' value={data.auto_certify ? 'Enabled' : 'Disabled'} />
            </div>
            <div className='mt-4'>
                <DetailField label='Selector Status' value={data.disabled_at ? 'Enabled' : 'Disabled'} />
            </div>
        </Card>
    );
};

const DynamicDetails: FC<DynamicDetailsProps> = ({ data }) => {
    if (isTag(data)) {
        return <TagDetails data={data} />;
    } else if (isSelector(data)) {
        return <SelectorDetails data={data} />;
    } else {
        return null;
    }
};

export default DynamicDetails;
