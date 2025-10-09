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

import { IconName } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Card, Skeleton } from 'doodle-ui';
import {
    AssetGroupTag,
    AssetGroupTagSelector,
    AssetGroupTagSelectorAutoCertifyMap,
    AssetGroupTagTypeZone,
    SeedTypeCypher,
    SeedTypesMap,
} from 'js-client-library';
import { DateTime } from 'luxon';
import { FC, useContext } from 'react';
import { UseQueryResult } from 'react-query';
import { useHighestPrivilegeTagId, useOwnedTagId, usePZPathParams } from '../../../hooks';
import { LuxonFormat } from '../../../utils';
import { Cypher } from '../Cypher/Cypher';
import { PrivilegeZonesContext } from '../PrivilegeZonesContext';
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

const TagDetails: FC<{ tagData: AssetGroupTag }> = ({ tagData }) => {
    const {
        glyph,
        name,
        description,
        position,
        created_by,
        updated_by,
        updated_at,
        id: tagId,
        type,
        require_certify,
    } = tagData;

    const lastUpdated = DateTime.fromISO(updated_at).toFormat(LuxonFormat.YEAR_MONTH_DAY_SLASHES);

    const { SalesMessage } = useContext(PrivilegeZonesContext);

    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const ownedId = useOwnedTagId();

    return (
        <div
            className='max-h-full flex flex-col gap-8 max-w-[32rem] w-full'
            data-testid='privilege-zones_tag-details-card'>
            <Card className='px-6 py-6'>
                <div className='text-xl font-bold truncate' title={name}>
                    {glyph && (
                        <span>
                            <FontAwesomeIcon icon={glyph as IconName} size='sm' /> <span> </span>
                        </span>
                    )}
                    {name}
                </div>
                {position !== null && (
                    <div className='mt-4'>
                        <DetailField label='Position' value={position.toString()} />
                    </div>
                )}
                <div className='mt-4'>
                    <DescriptionField description={description} />
                </div>
                <div className='mt-4'>
                    <DetailField label='Created By' value={created_by} />
                </div>
                <div className='mt-4'>
                    <DetailField label='Last Updated By' value={updated_by} />
                    <DetailField label='Last Updated' value={lastUpdated} />
                </div>
                {type === AssetGroupTagTypeZone && (
                    <div className='mt-4'>
                        <DetailField label='Certification' value={require_certify ? 'Required' : 'Not Required'} />
                    </div>
                )}
            </Card>
            {tagId !== topTagId && tagId !== ownedId && SalesMessage && <SalesMessage />}
            <ObjectCountPanel tagId={tagId.toString()} />
        </div>
    );
};

const SelectorDetails: FC<{ selectorData: AssetGroupTagSelector }> = ({ selectorData }) => {
    const { name, description, created_by, updated_by, updated_at, auto_certify, disabled_at, seeds } = selectorData;

    const lastUpdated = DateTime.fromISO(updated_at).toFormat(LuxonFormat.YEAR_MONTH_DAY_SLASHES);

    const seedType = getSelectorSeedType(selectorData);

    const { isZonePage } = usePZPathParams();

    return (
        <div
            className='max-h-full flex flex-col gap-8 max-w-[32rem]'
            data-testid='privilege-zones_selector-details-card'>
            <Card className='px-6 py-6'>
                <div className='text-xl font-bold truncate' title={name}>
                    {name}
                </div>
                <div className='mt-4'>
                    <DescriptionField description={description} />
                </div>
                <div className='mt-4'>
                    <DetailField label='Created By' value={created_by} />
                </div>
                <div className='mt-4'>
                    <DetailField label='Last Updated By' value={updated_by} />
                    <DetailField label='Last Updated' value={lastUpdated} />
                </div>
                {isZonePage && (
                    <div className='mt-4'>
                        <DetailField
                            label='Automatic Certification'
                            value={AssetGroupTagSelectorAutoCertifyMap[auto_certify] ?? 'Off'}
                        />
                    </div>
                )}

                <div className='mt-4'>
                    <DetailField label='Type' value={SeedTypesMap[seedType]} />
                    <DetailField label='Selector Status' value={disabled_at ? 'Disabled' : 'Enabled'} />
                </div>
            </Card>
            {seedType === SeedTypeCypher && <Cypher preview initialInput={seeds[0].value} />}
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
        return <TagDetails tagData={data} />;
    } else if (isSelector(data)) {
        return <SelectorDetails selectorData={data} />;
    }
    return null;
};

export default DynamicDetails;
