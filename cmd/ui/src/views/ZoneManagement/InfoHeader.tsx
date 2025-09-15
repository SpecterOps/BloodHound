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
import { Button } from '@bloodhoundenterprise/doodleui';
import {
    AD_PLATFORM,
    AZ_PLATFORM,
    AppLink,
    EnvironmentAggregation,
    privilegeZonesPath,
    selectorPath,
    savePath,
    SelectedEnvironment,
    SelectorValueTypes,
    SimpleEnvironmentSelector,
    getTagUrlValue,
    useEnvironmentParams,
    useHighestPrivilegeTagId,
    useInitialEnvironment,
} from 'bh-shared-ui';
import { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';

const aggregationFromType = (type: SelectorValueTypes | null): EnvironmentAggregation | null => {
    switch (type) {
        case AD_PLATFORM:
            return 'active-directory';
        case AZ_PLATFORM:
            return 'azure';
        default:
            return null;
    }
};

const InfoHeader: FC = () => {
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const { zoneId = topTagId?.toString(), labelId } = useParams();
    const tagId = labelId === undefined ? zoneId : labelId;

    const { data: initialEnvironment } = useInitialEnvironment({ orderBy: 'name' });

    const [selectedEnvironment, setSelectedEnvironment] = useState<SelectedEnvironment | undefined>(initialEnvironment);

    const { setEnvironmentParams } = useEnvironmentParams();

    const handleSelect = (environment: SelectedEnvironment) => {
        const { id, type } = environment;

        const aggregation = aggregationFromType(type);

        setEnvironmentParams({ environmentId: id, environmentAggregation: aggregation });

        setSelectedEnvironment(environment);
    };

    useEffect(() => {
        initialEnvironment && setSelectedEnvironment(initialEnvironment);
    }, [initialEnvironment]);

    return (
        <div className='flex justify-around basis-2/3'>
            <div className='flex justify-start gap-4 items-center basis-2/3'>
                <SimpleEnvironmentSelector
                    selected={{
                        type: selectedEnvironment?.type ?? null,
                        id: selectedEnvironment?.id ?? null,
                    }}
                    onSelect={handleSelect}
                />
                <Button variant='primary' disabled={!tagId} asChild>
                    <AppLink
                        data-testid='zone-management_create-selector-link'
                        to={`/${privilegeZonesPath}/${getTagUrlValue(labelId)}/${tagId}/${selectorPath}/${savePath}`}>
                        Create Selector
                    </AppLink>
                </Button>
            </div>
        </div>
    );
};

export default InfoHeader;
