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
    SelectedEnvironment,
    SelectorValueTypes,
    SimpleEnvironmentSelector,
    useEnvironmentParams,
    useHighestPrivilegeTagId,
    useInitialEnvironment,
    usePZPathParams,
} from 'bh-shared-ui';
import { FC, useEffect, useState } from 'react';

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
    const { tagId: defaultTagId, ruleCreateLink } = usePZPathParams();
    const tagId = !defaultTagId ? topTagId : defaultTagId;
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
        <div className='flex justify-start gap-4 items-center'>
            <SimpleEnvironmentSelector
                selected={{
                    type: selectedEnvironment?.type ?? null,
                    id: selectedEnvironment?.id ?? null,
                }}
                onSelect={handleSelect}
            />
            <Button variant='primary' disabled={!tagId} asChild>
                {!tagId ? (
                    'Create Rule'
                ) : (
                    <AppLink data-testid='privilege-zones_create-rule-link' to={ruleCreateLink(tagId)}>
                        Create Rule
                    </AppLink>
                )}
            </Button>
        </div>
    );
};

export default InfoHeader;
