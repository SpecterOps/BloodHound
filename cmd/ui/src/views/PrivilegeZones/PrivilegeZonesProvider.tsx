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
import {
    AD_PLATFORM,
    AZ_PLATFORM,
    EnvironmentAggregation,
    PrivilegeZonesContext,
    SelectedEnvironment,
    SelectorValueTypes,
    SimpleEnvironmentSelector,
    defaultPrivilegeZoneCtxValue,
    useEnvironmentParams,
    useInitialEnvironment,
} from 'bh-shared-ui';
import { useEffect, useState } from 'react';
import InfoHeader from './InfoHeader';

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

const EnvironmentSelector = () => {
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
        <SimpleEnvironmentSelector
            selected={{
                type: selectedEnvironment?.type ?? null,
                id: selectedEnvironment?.id ?? null,
            }}
            onSelect={handleSelect}
        />
    );
};

const PrivilegeZonesProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    return (
        <PrivilegeZonesContext.Provider value={{ ...defaultPrivilegeZoneCtxValue, EnvironmentSelector, InfoHeader }}>
            {children}
        </PrivilegeZonesContext.Provider>
    );
};

export default PrivilegeZonesProvider;
