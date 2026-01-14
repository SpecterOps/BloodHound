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
import { useHighestPrivilegeTagId, useOwnedTagId, usePZPathParams, usePrivilegeZoneAnalysis } from '../../../hooks';
import { LuxonFormat } from '../../../utils';
import { Cypher } from '../Cypher/Cypher';
import { PrivilegeZonesContext } from '../PrivilegeZonesContext';
import { ZoneIcon } from '../ZoneIcon';
import { getRuleSeedType, isRule, isTag } from '../utils';
import ObjectCountPanel from './ObjectCountPanel';

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

const TagDetails: FC<{ tagData: AssetGroupTag; hasObjectCountPanel: boolean }> = ({ tagData, hasObjectCountPanel }) => {
    const {
        glyph,
        name,
        description,
        created_by,
        updated_by,
        updated_at,
        id: tagId,
        type,
        require_certify,
        analysis_enabled,
    } = tagData;

    const lastUpdated = DateTime.fromISO(updated_at).toFormat(LuxonFormat.YEAR_MONTH_DAY_SLASHES);
    const { SalesMessage, Certification } = useContext(PrivilegeZonesContext);
    const privilegeZoneAnalysisEnabled = usePrivilegeZoneAnalysis();
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const ownedId = useOwnedTagId();

    return (
        <div className='max-h-full flex flex-col gap-8' data-testid='privilege-zones_tag-details-card'>
            <Card className='px-6 py-6 rounded-lg max-w-[32rem]'>
                <div className='flex items-center' title={name}>
                    {glyph && <ZoneIcon zone={tagData} persistGlyph size={20} />}
                    <span className='text-xl font-bold truncate'>{name}</span>
                </div>
                {Certification && (
                    <div className='mt-4'>
                        <DetailField
                            label='Analysis'
                            value={
                                (privilegeZoneAnalysisEnabled && analysis_enabled) || tagId === topTagId
                                    ? 'Enabled'
                                    : 'Disabled'
                            }
                        />
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
                {type === AssetGroupTagTypeZone && Certification && (
                    <div className='mt-4'>
                        <DetailField label='Certification' value={require_certify ? 'Required' : 'Not Required'} />
                    </div>
                )}
            </Card>
            {tagId !== topTagId && tagId !== ownedId && SalesMessage && <SalesMessage />}
            {hasObjectCountPanel && <ObjectCountPanel />}
        </div>
    );
};

const RuleDetails: FC<{ ruleData: AssetGroupTagSelector }> = ({ ruleData }) => {
    const { name, description, created_by, updated_by, updated_at, auto_certify, disabled_at, seeds } = ruleData;

    const lastUpdated = DateTime.fromISO(updated_at).toFormat(LuxonFormat.YEAR_MONTH_DAY_SLASHES);

    const seedType = getRuleSeedType(ruleData);

    const { isZonePage } = usePZPathParams();
    const { Certification } = useContext(PrivilegeZonesContext);

    return (
        <div className='max-h-full flex flex-col gap-8' data-testid='privilege-zones_selector-details-card'>
            <Card className='px-6 py-6 rounded-lg'>
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
                {isZonePage && Certification && (
                    <div className='mt-4'>
                        <DetailField
                            label='Automatic Certification'
                            value={AssetGroupTagSelectorAutoCertifyMap[auto_certify] ?? 'Off'}
                        />
                    </div>
                )}

                <div className='mt-4'>
                    <DetailField label='Type' value={SeedTypesMap[seedType]} />
                    <DetailField label='Rule Status' value={disabled_at ? 'Disabled' : 'Enabled'} />
                </div>
            </Card>
            {seedType === SeedTypeCypher && <Cypher preview initialInput={seeds[0].value} />}
        </div>
    );
};

type DynamicDetailsProps = {
    queryResult: UseQueryResult<AssetGroupTag | undefined> | UseQueryResult<AssetGroupTagSelector | undefined>;
    hasObjectCountPanel?: boolean;
};

const DynamicDetails: FC<DynamicDetailsProps> = ({
    queryResult: { isError, isLoading, data },
    hasObjectCountPanel = false,
}) => {
    if (isLoading) {
        return <Skeleton className='px-6 py-6 max-w-[32rem] h-52' />;
    } else if (isError) {
        return (
            <Card className='px-6 py-6 max-w-[32rem]'>
                <span className='text-base'>There was an error fetching this data</span>
            </Card>
        );
    } else if (isTag(data)) {
        return <TagDetails tagData={data} hasObjectCountPanel={hasObjectCountPanel} />;
    } else if (isRule(data)) {
        return <RuleDetails ruleData={data} />;
    }
    return null;
};

export default DynamicDetails;
