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

import {
    faCertificate,
    faCircleQuestion,
    faCog,
    faCube,
    faFile,
    faKey,
    faRoute,
    faShieldHalved,
    faTag,
    faTriangleExclamation,
    faUser,
    faUsers,
    IconDefinition,
} from '@fortawesome/free-solid-svg-icons';

export const RACF_NODE_KINDS = {
    User: 'RACFUser',
    Group: 'RACFGroup',
    Dataset: 'RACFDataset',
    Resource: 'RACFResource',
    Privilege: 'RACFPrivilege',
    Class: 'RACFClass',
    StartedTask: 'RACFStartedTask',
    Certificate: 'RACFCertificate',
    MFAFactor: 'RACFMFAFactor',
    Finding: 'RACFFinding',
    Path: 'RACFPath',
    Undefined: 'RACFUndefined',
} as const;

export type RACFNodeKind = (typeof RACF_NODE_KINDS)[keyof typeof RACF_NODE_KINDS];

type RACFIconInfo = {
    icon: IconDefinition;
    color: string;
};

const RACF_LEGACY_NODE_ICONS: Record<RACFNodeKind, RACFIconInfo> = {
    [RACF_NODE_KINDS.User]: {
        icon: faUser,
        color: '#4A90D9',
    },
    [RACF_NODE_KINDS.Group]: {
        icon: faUsers,
        color: '#58D68D',
    },
    [RACF_NODE_KINDS.Dataset]: {
        icon: faFile,
        color: '#F0B429',
    },
    [RACF_NODE_KINDS.Resource]: {
        icon: faCube,
        color: '#AF7AC5',
    },
    [RACF_NODE_KINDS.Privilege]: {
        icon: faKey,
        color: '#E74C3C',
    },
    [RACF_NODE_KINDS.Class]: {
        icon: faTag,
        color: '#1ABC9C',
    },
    [RACF_NODE_KINDS.StartedTask]: {
        icon: faCog,
        color: '#85929E',
    },
    [RACF_NODE_KINDS.Certificate]: {
        icon: faCertificate,
        color: '#D4AC0D',
    },
    [RACF_NODE_KINDS.MFAFactor]: {
        icon: faShieldHalved,
        color: '#2E86C1',
    },
    [RACF_NODE_KINDS.Finding]: {
        icon: faTriangleExclamation,
        color: '#E67E22',
    },
    [RACF_NODE_KINDS.Path]: {
        icon: faRoute,
        color: '#EC7063',
    },
    [RACF_NODE_KINDS.Undefined]: {
        icon: faCircleQuestion,
        color: '#B2BABB',
    },
};

// RACFHound currently emits legacy kinds while its modern OpenGraph contract
// uses the racf_ namespace. Keep both aliases visually identical during the
// migration.
export const RACF_NODE_ICONS: Record<string, RACFIconInfo> = Object.fromEntries(
    Object.entries(RACF_LEGACY_NODE_ICONS).flatMap(([kind, iconInfo]) => [
        [kind, iconInfo],
        [`racf_${kind}`, iconInfo],
    ])
);
