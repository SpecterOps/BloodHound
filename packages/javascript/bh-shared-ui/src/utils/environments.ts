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

import { IconDefinition, faCircleNodes, faCloud, faGlobe } from '@fortawesome/free-solid-svg-icons';
import { Environment, KnownEnvironmentType, knownEnvironmentTypes } from 'js-client-library';
import omit from 'lodash/omit';
import pick from 'lodash/pick';
import { EnvironmentAggregation, environmentAggregationMap } from '../hooks/useEnvironmentParams/useEnvironmentParams';
import { MappedStringLiteral } from '../types';
import { inRange } from './number';

export interface EnvironmentFilterCheckboxState extends MappedStringLiteral<Environment['type'], boolean> {
    critical: boolean;
    high: boolean;
    moderate: boolean;
    low: boolean;
    yes: boolean;
    no: boolean;
}

export type EnvironmentInfo = {
    aggregationDisplayName: string;
    displayName: string;
    icon: IconDefinition;
    memberType: string;
    type: Environment['type'];
};

export const DEFAULT_ENVIRONMENTS_FILTER = {
    'active-directory': false,
    azure: false,
    critical: false,
    high: false,
    moderate: false,
    low: false,
    yes: false,
    no: false,
} satisfies EnvironmentFilterCheckboxState;

// Risk level thresholds
export const CRITICAL_THRESHOLD = 95;
export const HIGH_THRESHOLD = 80;
export const MODERATE_THRESHOLD = 40;

export const noEnvironmentSelectedFallback = 'No Environment Selected';

export const knownEnvironmentInfoMap: Record<KnownEnvironmentType, EnvironmentInfo> = {
    'active-directory': {
        aggregationDisplayName: 'All Active Directory Domains',
        displayName: 'Active Directory',
        icon: faGlobe,
        memberType: 'Domain',
        type: 'active-directory',
    },
    azure: {
        aggregationDisplayName: 'All Azure Tenants',
        displayName: 'Azure',
        icon: faCloud,
        memberType: 'Tenant',
        type: 'azure',
    },
};

const collectedFilterKeys = ['yes', 'no'] as const;
const collectedKeyBoolMap = {
    yes: true,
    no: false,
} as const;

const riskFilterKeys = ['critical', 'high', 'moderate', 'low'] as const;

export function filterAndSearchEnvironments(
    environments: Environment[] = [],
    options: {
        filters?: Partial<EnvironmentFilterCheckboxState>;
        search?: string;
    } = {}
) {
    const { search = '' } = options;
    const filters = { ...DEFAULT_ENVIRONMENTS_FILTER, ...options.filters };

    if (environments.length === 0) return environments;

    // Collected filter applies if at least one selected but not if all are selected
    const collectedFilters = pick(filters, collectedFilterKeys);
    let considerCollected = Object.values(collectedFilters).some((value) => value);
    considerCollected &&= !Object.values(collectedFilters).every((value) => value);

    // Risk filter applies if at least one selected but not if all are selected
    const riskFilters = pick(filters, riskFilterKeys);
    let considerRisk = Object.values(riskFilters).some((value) => value);
    considerRisk &&= !Object.values(riskFilters).every((value) => value);

    // Everything else is a platform filter which may be a known type or user provided
    const platformFilters = omit(filters, [...collectedFilterKeys, ...riskFilterKeys]);
    const considerPlatform = Object.values(platformFilters).some((value) => value);

    return environments.filter(({ collected, impactValue, name, type }) => {
        let isIncluded = true;

        if (considerCollected) {
            // Keep if collected status matches a selected filter
            isIncluded &&= Object.entries(collectedFilters).some(
                ([filterKey, enabled]) =>
                    enabled && collected === collectedKeyBoolMap[filterKey as keyof typeof collectedKeyBoolMap]
            );
        }

        if (considerRisk) {
            // Discard result without checking risk for non collected environments since they have no severity value
            isIncluded &&= collected;

            const { critical, high, moderate, low } = filters;

            // Keep if impact value is within filtered risk level thresholds
            isIncluded &&=
                ((critical && impactValue > CRITICAL_THRESHOLD) ||
                    (high && inRange(impactValue, HIGH_THRESHOLD, CRITICAL_THRESHOLD)) ||
                    (moderate && inRange(impactValue, MODERATE_THRESHOLD, HIGH_THRESHOLD)) ||
                    (low && MODERATE_THRESHOLD >= impactValue)) ??
                isIncluded;
        }

        // Only keep if environment matches a selected platform filter
        if (considerPlatform) {
            isIncluded &&= Object.entries(platformFilters).some(
                ([filterKey, enabled]) => enabled && filterKey === type
            );
        }

        // Finally, only keep if environment name match search
        if (search) {
            isIncluded &&= name.toLowerCase().includes(search.toLowerCase());
        }

        return isIncluded;
    });
}

export function getCheckboxOptions(environmentMap: Record<Environment['type'], { displayName: string }>) {
    return Object.entries(environmentMap).map(([name, { displayName }]) => ({
        name,
        label: displayName,
    }));
}

/** Return an object containing display name, aggregation name, member type, and icon for a given environment type */
export function getOpenGraphEnvironmentInfo(type: Environment['type']): EnvironmentInfo {
    // Known types (AD and Azure) use the known info map
    // Defaults used for OpenGraph types
    const { aggregationDisplayName, displayName, icon, memberType } = knownEnvironmentInfoMap[
        type as KnownEnvironmentType
    ] ?? {
        aggregationDisplayName: `All ${type} Environments`,
        displayName: type,
        icon: faCircleNodes,
        memberType: 'Name',
    };
    return {
        aggregationDisplayName,
        displayName,
        icon,
        memberType,
        type,
    };
}

/** Return a map of environment types to their display name, aggregation name, member type, and icon */
export function getOpenGraphEnvironmentInfoMap(environments: Environment[] = []) {
    const knownEnvironmentInfoCopy = { ...(knownEnvironmentInfoMap as Record<Environment['type'], EnvironmentInfo>) };
    if (environments === null) return knownEnvironmentInfoCopy;

    return environments.reduce(
        (acc, { type }) => {
            // Map starts with known types (AD and Azure)
            // OpenGraph types are added dynamically
            if (!acc[type]) {
                acc[type] = getOpenGraphEnvironmentInfo(type);
            }
            return acc;
        },
        { ...knownEnvironmentInfoCopy }
    );
}

export function isEnvironmentAggregation(id: string): id is EnvironmentAggregation {
    return Object.prototype.hasOwnProperty.call(environmentAggregationMap, id);
}

export function isKnownEnvironmentType(type?: string): type is KnownEnvironmentType {
    return knownEnvironmentTypes.includes(type as KnownEnvironmentType);
}

export function sortEnvironmentsByName(a: Environment, b: Environment) {
    return a.name.localeCompare(b.name);
}
