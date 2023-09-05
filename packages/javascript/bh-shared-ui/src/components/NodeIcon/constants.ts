// Copyright 2023 Specter Ops, Inc.
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
    faBox,
    faBoxOpen,
    faBolt,
    faClipboardList,
    faCloud,
    faCube,
    faCubes,
    faDesktop,
    faCog,
    faGlobe,
    faKey,
    faList,
    faLock,
    faObjectGroup,
    faRobot,
    faServer,
    faSitemap,
    faUser,
    faUsers,
    faWindowRestore,
    IconDefinition,
} from '@fortawesome/free-solid-svg-icons';

export const NODE_ICON: { [index: string]: { icon: IconDefinition; color: string } } = {
    User: {
        icon: faUser,
        color: '#17E625',
    },

    Group: {
        icon: faUsers,
        color: '#DBE617',
    },

    Computer: {
        icon: faDesktop,
        color: '#E67873',
    },

    Domain: {
        icon: faGlobe,
        color: '#17E6B9',
    },

    GPO: {
        icon: faList,
        color: '#998EFD',
    },

    OU: {
        icon: faSitemap,
        color: '#FFAA00',
    },

    Container: {
        icon: faBox,
        color: '#F79A78',
    },

    AZUser: {
        icon: faUser,
        color: '#34D2EB',
    },

    AZGroup: {
        icon: faUsers,
        color: '#F57C9B',
    },

    AZTenant: {
        icon: faCloud,
        color: '#54F2F2',
    },

    AZSubscription: {
        icon: faKey,
        color: '#D2CCA1',
    },

    AZResourceGroup: {
        icon: faCube,
        color: '#89BD9E',
    },

    AZVM: {
        icon: faDesktop,
        color: '#F9ADA0',
    },
    AZWebApp: {
        icon: faObjectGroup,
        color: '#4696E9',
    },
    AZLogicApp: {
        icon: faSitemap,
        color: '#9EE047',
    },

    AZAutomationAccount: {
        icon: faCog,
        color: '#F4BA44',
    },

    AZFunctionApp: {
        icon: faBolt,
        color: '#F4BA44',
    },

    AZContainerRegistry: {
        icon: faBoxOpen,
        color: '#0885D7',
    },

    AZManagedCluster: {
        icon: faCubes,
        color: '#326CE5',
    },

    AZDevice: {
        icon: faDesktop,
        color: '#B18FCF',
    },

    AZKeyVault: {
        icon: faLock,
        color: '#ED658C',
    },

    AZApp: {
        icon: faWindowRestore,
        color: '#03FC84',
    },

    AZVMScaleSet: {
        icon: faServer,
        color: '#007CD0',
    },

    AZServicePrincipal: {
        icon: faRobot,
        color: '#C1D6D6',
    },

    AZRole: {
        icon: faClipboardList,
        color: '#ED8537',
    },

    AZManagementGroup: {
        icon: faSitemap,
        color: '#BD93D8',
    },
};
