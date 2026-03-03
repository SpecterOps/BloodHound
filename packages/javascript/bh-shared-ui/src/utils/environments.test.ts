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

import { type Environment } from 'js-client-library';
import { getEnvironmentAggregationIds } from '../hooks/useEnvironmentIdList';
import { testEnvironments } from '../mocks/handlers/environments';
import {
    CRITICAL_THRESHOLD,
    DEFAULT_ENVIRONMENTS_FILTER,
    filterAndSearchEnvironments,
    HIGH_THRESHOLD,
    isEnvironmentAggregation,
    MODERATE_THRESHOLD,
} from './environments';

describe('filterAndSearchEnvironments', () => {
    describe('with empty environments', () => {
        it('returns empty array when environments is empty', () => {
            const result = filterAndSearchEnvironments([]);
            expect(result).toEqual([]);
        });

        it('returns empty array when environments is undefined', () => {
            const result = filterAndSearchEnvironments(undefined);
            expect(result).toEqual([]);
        });
    });

    describe('with no filters applied', () => {
        it('returns all environments when no filters are active', () => {
            const result = filterAndSearchEnvironments(testEnvironments);
            expect(result).toEqual(testEnvironments);
        });
    });

    describe('collected filters', () => {
        it('filters by collected=true when "yes" filter is active', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { yes: true },
            });
            expect(result).toHaveLength(10);
            expect(result.every((env) => env.collected)).toBe(true);
        });

        it('filters by collected=false when "no" filter is active', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { no: true },
            });
            expect(result).toHaveLength(5);
            expect(result.every((env) => !env.collected)).toBe(true);
        });

        it('returns all environments when both "yes" and "no" filters are active', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { yes: true, no: true },
            });
            expect(result).toEqual(testEnvironments);
        });
    });

    describe('risk filters', () => {
        it('filters by critical risk level', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { critical: true },
            });
            expect(result).toHaveLength(2);
            expect(result[0].impactValue).toBeGreaterThan(CRITICAL_THRESHOLD);
            expect(result[0].collected).toBe(true);
        });

        it('filters by high risk level', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { high: true },
            });
            expect(result).toHaveLength(2);
            expect(result.every((env) => env.impactValue > HIGH_THRESHOLD)).toBe(true);
            expect(result.every((env) => env.impactValue <= CRITICAL_THRESHOLD)).toBe(true);
            expect(result.every((env) => env.collected)).toBe(true);
        });

        it('filters by moderate risk level', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { moderate: true },
            });
            expect(result).toHaveLength(3);
            expect(result[0].impactValue).toBeGreaterThan(MODERATE_THRESHOLD);
            expect(result[0].impactValue).toBeLessThanOrEqual(HIGH_THRESHOLD);
            expect(result[0].collected).toBe(true);
        });

        it('filters by low risk level', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { low: true },
            });
            expect(result).toHaveLength(3);
            expect(result[0].impactValue).toBeLessThanOrEqual(MODERATE_THRESHOLD);
            expect(result[0].collected).toBe(true);
        });

        it('filters by multiple risk levels', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { critical: true, high: true },
            });
            expect(result).toHaveLength(4);
            expect(result.every((env) => env.impactValue > HIGH_THRESHOLD)).toBe(true);
        });

        it('excludes non-collected environments when risk filter is active', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { critical: true },
            });
            expect(result.every((env) => env.collected)).toBe(true);
        });
    });

    describe('platform filters', () => {
        it('filters by azure platform', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { azure: true },
            });
            expect(result).toHaveLength(9);
            expect(result.every((env) => env.type === 'azure')).toBe(true);
        });

        it('filters by active-directory platform', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { 'active-directory': true },
            });
            expect(result).toHaveLength(3);
            expect(result.every((env) => env.type === 'active-directory')).toBe(true);
        });

        it('filters by user-provided platform', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { GitHub: true },
            });
            expect(result).toHaveLength(2);
            expect(result.every((env) => env.type === 'GitHub')).toBe(true);
        });

        it('filters by multiple platforms', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { azure: true, AWS: true },
            });
            expect(result).toHaveLength(10);
        });
    });

    describe('search filter', () => {
        it('filters by name search (case insensitive)', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: DEFAULT_ENVIRONMENTS_FILTER,
                search: '.Net',
            });
            expect(result).toHaveLength(3);

            const names = result.map((env) => env.name);
            expect(names).toContain('clementina.net');
            expect(names).toContain('adan.net');
            expect(names).toContain('pros.net');
        });

        it('returns empty array when search does not match any environment', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: DEFAULT_ENVIRONMENTS_FILTER,
                search: 'nonexistent',
            });
            expect(result).toEqual([]);
        });

        it('handles empty search string', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: DEFAULT_ENVIRONMENTS_FILTER,
                search: '',
            });
            expect(result).toEqual(testEnvironments);
        });
    });

    describe('combined filters', () => {
        it('combines collected and risk filters', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { yes: true, critical: true },
            });
            expect(result).toHaveLength(2);
            expect(result[0].collected).toBe(true);
            expect(result[0].impactValue).toBeGreaterThan(CRITICAL_THRESHOLD);
        });

        it('combines platform and risk filters', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { azure: true, high: true },
            });
            expect(result).toHaveLength(2);
            expect(result[0].type).toBe('azure');
            expect(result[0].impactValue).toBeGreaterThan(HIGH_THRESHOLD);
            expect(result[0].impactValue).toBeLessThanOrEqual(CRITICAL_THRESHOLD);
        });

        it('combines platform and collected filters', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { 'active-directory': true, no: true },
            });
            expect(result).toHaveLength(2);
            expect(result[0].type).toBe('active-directory');
            expect(result[0].collected).toBe(false);
        });

        it('combines all filter types', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                search: 'dorothea',
                filters: { azure: true, yes: true, critical: true },
            });
            expect(result).toHaveLength(1);
            expect(result[0].type).toBe('azure');
            expect(result[0].collected).toBe(true);
            expect(result[0].impactValue).toBeGreaterThan(CRITICAL_THRESHOLD);
            expect(result[0].name).toContain('dorothea');
        });

        it('returns empty array when combined filters match nothing', () => {
            const result = filterAndSearchEnvironments(testEnvironments, {
                filters: { 'active-directory': true, critical: true },
            });
            expect(result).toEqual([]);
        });
    });

    describe('edge cases', () => {
        it('handles environment with impactValue at exact threshold boundaries', () => {
            const boundaryEnvs: Environment[] = [
                {
                    type: 'azure',
                    impactValue: CRITICAL_THRESHOLD,
                    name: 'at-critical-threshold',
                    id: '1',
                    collected: true,
                    hygiene_attack_paths: 0,
                    exposures: [],
                },
                {
                    type: 'azure',
                    impactValue: HIGH_THRESHOLD,
                    name: 'at-high-threshold',
                    id: '2',
                    collected: true,
                    hygiene_attack_paths: 0,
                    exposures: [],
                },
                {
                    type: 'azure',
                    impactValue: MODERATE_THRESHOLD,
                    name: 'at-moderate-threshold',
                    id: '3',
                    collected: true,
                    hygiene_attack_paths: 0,
                    exposures: [],
                },
            ];

            const criticalResult = filterAndSearchEnvironments(boundaryEnvs, {
                filters: { critical: true },
            });
            expect(criticalResult).toHaveLength(0); // > CRITICAL_THRESHOLD

            const highResult = filterAndSearchEnvironments(boundaryEnvs, {
                filters: { high: true },
            });
            expect(highResult).toHaveLength(1); // > HIGH_THRESHOLD && <= CRITICAL_THRESHOLD
            expect(highResult[0].impactValue).toBe(CRITICAL_THRESHOLD);

            const moderateResult = filterAndSearchEnvironments(boundaryEnvs, {
                filters: { moderate: true },
            });
            expect(moderateResult).toHaveLength(1); // > MODERATE_THRESHOLD && <= HIGH_THRESHOLD
            expect(moderateResult[0].impactValue).toBe(HIGH_THRESHOLD);

            const lowResult = filterAndSearchEnvironments(boundaryEnvs, {
                filters: { low: true },
            });
            expect(lowResult).toHaveLength(1); // <= MODERATE_THRESHOLD
            expect(lowResult[0].impactValue).toBe(MODERATE_THRESHOLD);
        });

        it('handles environment with impactValue of 0', () => {
            const zeroImpactEnv: Environment[] = [
                {
                    type: 'azure',
                    impactValue: 0,
                    name: 'zero-impact',
                    id: '1',
                    collected: true,
                    hygiene_attack_paths: 0,
                    exposures: [],
                },
            ];

            const result = filterAndSearchEnvironments(zeroImpactEnv, {
                filters: { low: true },
            });
            expect(result).toHaveLength(1);
        });
    });
});

describe('getEnvironmentAggregationIds', () => {
    it('returns all collected environment ids associated to a given aggregation', () => {
        const expected = [
            '031664c0-a9e4-4355-8c27-a43994619cc4',
            '142775d1-a9e4-4355-9d38-b34885728dd3',
            '2e307757-7dd1-5b52-ceb5-g69fd8c7fcd5',
            '3g318857-8bb1-5b74-cgb5-g67cb8c7gcd8',
            '8c5e163g-d3ce-4a8e-b02e-08857611ac67',
            '9d5f073f-d3ce-4a8e-b02e-08857611ac67',
            '9d6h153h-e4c6-4a8e-b02e-19947611bd68',
        ];
        const actual = getEnvironmentAggregationIds('azure', testEnvironments);
        expect(actual).toEqual(expected);
    });

    it('returns all collected environments when aggegation type is all', () => {
        const expected = [
            '031664c0-a9e4-4355-8c27-a43994619cc4',
            '0a64f3b5b-eee1-4262-a6f8-14133071a18e',
            '142775d1-a9e4-4355-9d38-b34885728dd3',
            '2e307757-7dd1-5b52-ceb5-g69fd8c7fcd5',
            '3a6f8001-11f4-43bb-9de6-25c0d931f244',
            '3g318857-8bb1-5b74-cgb5-g67cb8c7gcd8',
            '61cc44b36-bf1d-4fac-bd90-35ed2eecdf03',
            '8c5e163g-d3ce-4a8e-b02e-08857611ac67',
            '9d5f073f-d3ce-4a8e-b02e-08857611ac67',
            '9d6h153h-e4c6-4a8e-b02e-19947611bd68',
        ];
        const actual = getEnvironmentAggregationIds('all', testEnvironments);
        expect(actual).toEqual(expected);
    });
});

describe('isEnvironmentAggregation', () => {
    it('returns true when id is a environment aggregation type', () => {
        const expected = true;
        const actual = isEnvironmentAggregation('active-directory');
        expect(actual).toBe(expected);
    });
    it('returns false when id is not an environment aggregation', () => {
        const expected = false;
        const actual = isEnvironmentAggregation('123');
        expect(actual).toBe(expected);
    });
});
