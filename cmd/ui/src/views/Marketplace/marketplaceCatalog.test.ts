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
    communityExtensions,
    communityIntegrations,
    developmentExtensions,
    enterpriseExtensions,
    enterpriseIntegrations,
} from './marketplaceCatalog';

describe('Marketplace catalog', () => {
    const catalog = [
        ...enterpriseExtensions,
        ...communityExtensions,
        ...enterpriseIntegrations,
        ...communityIntegrations,
        ...developmentExtensions,
    ];

    it('contains complete, uniquely identified CE records', () => {
        expect(enterpriseExtensions).toHaveLength(5);
        expect(communityExtensions).toHaveLength(26);
        expect(enterpriseIntegrations).toHaveLength(11);
        expect(communityIntegrations).toHaveLength(3);
        expect(developmentExtensions).toHaveLength(8);
        expect(
            new Set(catalog.map(({ name, publisher, availability }) => `${name}:${publisher}:${availability}`)).size
        ).toBe(catalog.length);

        for (const item of catalog) {
            expect(item).toEqual(
                expect.objectContaining({
                    author: expect.any(String),
                    availability: expect.stringMatching(/^(general|early-access|coming-soon)$/),
                    description: expect.any(String),
                    href: expect.stringMatching(/^https:\/\//),
                    logoPath: expect.stringMatching(/^\/img\/.+\.(png|svg)$/),
                    name: expect.any(String),
                    publisher: expect.stringMatching(/^(specterops|community|partner)$/),
                })
            );
        }
    });

    it('models Enterprise availability and keeps the CE catalog free of Enterprise-only destinations', () => {
        expect(enterpriseExtensions.filter(({ availability }) => availability === 'early-access')).toEqual([
            expect.objectContaining({ name: 'AWS', badge: 'Beta' }),
            expect.objectContaining({ name: 'Microsoft Entra Agent ID', badge: 'Beta' }),
        ]);
        expect(enterpriseIntegrations.filter(({ publisher }) => publisher === 'partner')).toHaveLength(4);
        expect(enterpriseIntegrations.every(({ availability }) => availability === 'general')).toBe(true);
        expect(developmentExtensions).toEqual(
            expect.arrayContaining([
                expect.objectContaining({ name: 'AWS Bedrock', availability: 'coming-soon', publisher: 'specterops' }),
                expect.objectContaining({
                    name: 'Entra Domain Services',
                    availability: 'coming-soon',
                    publisher: 'specterops',
                }),
                expect.objectContaining({ name: 'Azure DevOps', availability: 'coming-soon', publisher: 'specterops' }),
                expect.objectContaining({
                    name: 'HashiCorp Vault',
                    availability: 'coming-soon',
                    publisher: 'specterops',
                }),
                expect.objectContaining({ name: 'Teleport', availability: 'coming-soon', publisher: 'specterops' }),
                expect.objectContaining({ name: 'GitLab', availability: 'coming-soon', publisher: 'specterops' }),
                expect.objectContaining({ name: 'Intune', availability: 'coming-soon', publisher: 'specterops' }),
                expect.objectContaining({ name: 'Snowflake', availability: 'coming-soon', publisher: 'specterops' }),
            ])
        );
        expect(
            developmentExtensions.every(({ href }) => href === 'https://pages.specterops.io/BHE-Extensions-Beta.html')
        ).toBe(true);

        for (const item of [...communityExtensions, ...communityIntegrations]) {
            expect(`${item.name} ${item.description}`).not.toMatch(/BloodHound Enterprise|upgrade your license/i);
            expect(item.href).not.toContain('/integrations/');
            expect(item.href).toMatch(/^https:\/\/github\.com\//);
            expect(item.availability).toBe('general');
        }
    });
});
