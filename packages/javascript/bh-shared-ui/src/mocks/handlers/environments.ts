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

import { Environment } from 'js-client-library';

export const testEnvironments: Environment[] = [
    {
        type: 'active-directory',
        impactValue: 54,
        name: 'eladio.info',
        id: '3a6f8001-11f4-43bb-9de6-25c0d931f244',
        collected: true,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'active-directory',
        impactValue: 64,
        name: 'adan.net',
        id: 'ab84177d-aac4-4923-bf0a-d279fd41462c',
        collected: false,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'active-directory',
        impactValue: 93,
        name: 'omer.com',
        id: '171adcb0-a3ee-4d57-8f74-f040e72b3890',
        collected: false,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'azure',
        impactValue: 84,
        name: 'clementina.net',
        id: '031664c0-a9e4-4355-8c27-a43994619cc4',
        collected: true,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'azure',
        impactValue: 97,
        name: 'dorothea.info',
        id: '9d5f073f-d3ce-4a8e-b02e-08857611ac67',
        collected: true,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'azure',
        impactValue: 95,
        name: 'blankspace.info',
        id: '9d6h153h-e4c6-4a8e-b02e-19947611bd68',
        collected: true,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'azure',
        impactValue: 80,
        name: 'cardigan.com',
        id: '142775d1-a9e4-4355-9d38-b34885728dd3',
        collected: true,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'azure',
        impactValue: 1,
        name: 'antihero.com',
        id: '2e307757-7dd1-5b52-ceb5-g69fd8c7fcd5',
        collected: true,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'azure',
        impactValue: 20,
        name: 'lavender.io',
        id: '3g318857-8bb1-5b74-cgb5-g67cb8c7gcd8',
        collected: true,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'azure',
        impactValue: 96,
        name: 'august.com',
        id: '8c5e163g-d3ce-4a8e-b02e-08857611ac67',
        collected: true,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'azure',
        impactValue: 40,
        name: 'maroon.info',
        id: '3d1088855-5bb9-3z52-adz3-d47bb6a5dac5',
        collected: false,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'azure',
        impactValue: 22,
        name: 'ophelia.biz',
        id: '1f209946-6cc0-4a63-bfa4-f58dc7b6ebc6',
        collected: false,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'AWS',
        impactValue: 30,
        name: 'cyan.info',
        id: '0a64f3b5b-eee1-4262-a6f8-14133071a18e',
        collected: true,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'GitHub',
        impactValue: 64,
        name: 'steve.code',
        id: '61cc44b36-bf1d-4fac-bd90-35ed2eecdf03',
        collected: true,
        hygiene_attack_paths: 0,
        exposures: [],
    },
    {
        type: 'GitHub',
        impactValue: 93,
        name: 'pros.net',
        id: '1894b90b6-5946-4bcd-9fd7-3451f669f95d',
        collected: false,
        hygiene_attack_paths: 0,
        exposures: [],
    },
];
