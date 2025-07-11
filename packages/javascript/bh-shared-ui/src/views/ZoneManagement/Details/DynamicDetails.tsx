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

import { Card, Skeleton } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTag, AssetGroupTagSelector, SeedTypeCypher, SeedTypesMap } from 'js-client-library';
import { DateTime } from 'luxon';
import { FC, useContext } from 'react';
import { UseQueryResult } from 'react-query';
import { useParams } from 'react-router-dom';
import { LuxonFormat } from '../../../utils';
import { Cypher } from '../Cypher/Cypher';
import { ZoneManagementContext } from '../ZoneManagementContext';
import { OWNED_ID, TIER_ZERO_ID } from '../utils';
import ObjectCountPanel from './ObjectCountPanel';
import { getSelectorSeedType, isSelector, isTag } from './utils';

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

const DescriptionField: FC<{ description: string }> = ({ description }) => {
    return (
        <div className='flex flex-col gap-x-2'>
            <span className='font-bold'>Description:</span>
            <div className='max-h-36 overflow-y-auto'>
                <p title={description}>{description}</p>
            </div>
        </div>
    );
};

const TagDetails: FC<{ data: AssetGroupTag }> = ({ data }) => {
    const lastUpdated = DateTime.fromISO(data.updated_at).toFormat(LuxonFormat.YEAR_MONTH_DAY_SLASHES);
    const { SalesMessage } = useContext(ZoneManagementContext);
    const { tierId = '', labelId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;

    return (
        <div className='max-h-full flex flex-col gap-8 max-w-[32rem]' data-testid='zone-management_tag-details-card'>
            <Card className='px-6 py-6'>
                <div className='text-xl font-bold truncate' title={data.name}>
                    {data.name}
                </div>
                {data.position !== null && (
                    <div className='mt-4'>
                        <DetailField label='Position' value={data.position.toString()} />
                    </div>
                )}
                <div className='mt-4'>
                    <DescriptionField description={data.description} />
                </div>
                <div className='mt-4'>
                    <DetailField label='Created By' value={data.created_by} />
                </div>
                <div className='mt-4'>
                    <DetailField label='Last Updated By' value={data.updated_by} />
                    <DetailField label='Last Updated' value={lastUpdated} />
                </div>
                <div className='mt-4' hidden>
                    <DetailField label='Certification' value={data.requireCertify ? 'Required' : 'Not Required'} />
                </div>
            </Card>
            {tagId !== TIER_ZERO_ID && tagId !== OWNED_ID && SalesMessage && <SalesMessage />}
            <ObjectCountPanel tagId={data.id.toString()} />
        </div>
    );
};

const SelectorDetails: FC<{ data: AssetGroupTagSelector }> = ({ data }) => {
    const lastUpdated = DateTime.fromISO(data.updated_at).toFormat(LuxonFormat.YEAR_MONTH_DAY_SLASHES);
    const seedType = getSelectorSeedType(data);

    return (
        <div
            className='max-h-full flex flex-col gap-8 max-w-[32rem]'
            data-testid='zone-management_selector-details-card'>
            <Card className='px-6 py-6'>
                <div className='text-xl font-bold truncate' title={data.name}>
                    {data.name}
                </div>
                <div className='mt-4'>
                    <DescriptionField description={data.description} />
                </div>
                <div className='mt-4'>
                    <DetailField label='Created By' value={data.created_by} />
                </div>
                <div className='mt-4'>
                    <DetailField label='Last Updated By' value={data.updated_by} />
                    <DetailField label='Last Updated' value={lastUpdated} />
                </div>
                <div className='mt-4' hidden>
                    <DetailField label='Automatic Certification' value={data.auto_certify ? 'Enabled' : 'Disabled'} />
                </div>
                <div className='mt-4'>
                    <DetailField label='Type' value={SeedTypesMap[seedType]} />
                    <DetailField label='Selector Status' value={data.disabled_at ? 'Disabled' : 'Enabled'} />
                </div>
            </Card>
            {getSelectorSeedType(data) === SeedTypeCypher && <Cypher preview initialInput={data.seeds[0].value} />}
        </div>
    );
};

type DynamicDetailsProps = {
    queryResult: UseQueryResult<AssetGroupTag | undefined> | UseQueryResult<AssetGroupTagSelector | undefined>;
};

const DynamicDetails: FC<DynamicDetailsProps> = ({ queryResult: { isError, isLoading, data } }) => {
    if (isLoading) {
        return <Skeleton className='px-6 py-6 max-w-[32rem] h-52' />;
    } else if (isError) {
        return (
            <Card className='px-6 py-6 max-w-[32rem]'>
                <span className='text-base'>There was an error fetching this data</span>
            </Card>
        );
    } else if (isTag(data)) {
        return <TagDetails data={data} />;
    } else if (isSelector(data)) {
        return <SelectorDetails data={data} />;
    }
    return null;
};

export default DynamicDetails;
