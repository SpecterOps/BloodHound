// Copyright 2026 Specter Ops, Inc.
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

import { SelectedEnvironment } from '../../../components/SimpleEnvironmentSelector/types';
import { ActiveDirectoryPlatformInfo, DomainInfo } from '../DomainInfo';
import OpenGraphInfo, { OpenGraphPlatformInfo } from '../OpenGraphInfo';
import TenantInfo, { AzurePlatformInfo } from '../TenantInfo';

export const getStatsComponent = (selectedEnvironment: SelectedEnvironment | null, dataErrorHandler: () => void) => {
    const contextType = selectedEnvironment?.type;
    const contextId = selectedEnvironment?.id;
    switch (contextType) {
        case 'active-directory':
            if (!contextId) return null;
            return <DomainInfo contextId={contextId} onDataError={dataErrorHandler} />;
        case 'active-directory-platform':
            return <ActiveDirectoryPlatformInfo onDataError={dataErrorHandler} />;
        case 'azure':
            if (!contextId) return null;
            return <TenantInfo contextId={contextId} onDataError={dataErrorHandler} />;
        case 'azure-platform':
            return <AzurePlatformInfo onDataError={dataErrorHandler} />;
        default:
            if (!contextType) return null;
            if (contextType.endsWith('-platform')) {
                return (
                    <OpenGraphPlatformInfo
                        contextType={contextType.replace('-platform', '')}
                        onDataError={dataErrorHandler}
                    />
                );
            }
            if (!contextId) return null;
            return <OpenGraphInfo contextId={contextId} onDataError={dataErrorHandler} />;
    }
};
