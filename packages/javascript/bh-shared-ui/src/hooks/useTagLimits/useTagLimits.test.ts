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

import { AssetGroupTagTypeLabel, ConfigurationKey } from 'js-client-library';
import { createAssetGroupTag } from '../../mocks/factories/privilegeZones';
import { renderHook } from '../../test-utils';
import { setUpQueryClient } from '../../utils';
import { privilegeZonesKeys } from '../useAssetGroupTags';
import { configurationKeys } from '../useConfiguration';
import { useTagLimits } from './useTagLimits';

describe('useTagLimits', () => {
    it('returns when zone limit is reached', () => {
        const MOCK_CONFIG = {
            data: [
                {
                    key: 'analysis.tiering',
                    name: 'Multi-Tier Analysis Configuration',
                    description:
                        'This configuration parameter determines the limits of tiering with respect to analysis',
                    value: {
                        //How many zones can be created?
                        tier_limit: 1,
                        label_limit: 10,
                        multi_tier_analysis_enabled: true,
                    },
                    id: 8,
                    created_at: '2025-11-03T17:03:28.42299Z',
                    updated_at: '2025-11-03T17:03:28.42299Z',
                    deleted_at: {
                        Time: '0001-01-01T00:00:00Z',
                        Valid: false,
                    },
                },
            ],
        };

        const mockState = [
            {
                key: privilegeZonesKeys.tags(),
                //Create as many zones as needed to match the tier_limit
                data: [createAssetGroupTag()],
            },
            { key: configurationKeys.all, data: MOCK_CONFIG },
            { key: ConfigurationKey.Tiering, data: MOCK_CONFIG },
        ];
        const queryClient = setUpQueryClient(mockState);

        const { result } = renderHook(() => useTagLimits(), { queryClient });
        // zoneLimitReached will be true when the amount of Zones created are equal to the tier_limit
        expect(result.current.zoneLimitReached).toBe(true);
    });

    it('does not return when zone limit is not reached', () => {
        const MOCK_CONFIG = {
            data: [
                {
                    key: 'analysis.tiering',
                    name: 'Multi-Tier Analysis Configuration',
                    description:
                        'This configuration parameter determines the limits of tiering with respect to analysis',
                    value: {
                        //How many zones can be created?
                        tier_limit: 3,
                        label_limit: 10,
                        multi_tier_analysis_enabled: true,
                    },
                    id: 8,
                    created_at: '2025-11-03T17:03:28.42299Z',
                    updated_at: '2025-11-03T17:03:28.42299Z',
                    deleted_at: {
                        Time: '0001-01-01T00:00:00Z',
                        Valid: false,
                    },
                },
            ],
        };

        const mockState = [
            {
                key: privilegeZonesKeys.tags(),
                //Create as many zones as needed without reaching tier_limit
                data: [createAssetGroupTag(), createAssetGroupTag()],
            },
            { key: configurationKeys.all, data: MOCK_CONFIG },
            { key: ConfigurationKey.Tiering, data: MOCK_CONFIG },
        ];
        const queryClient = setUpQueryClient(mockState);

        const { result } = renderHook(() => useTagLimits(), { queryClient });
        // zoneLimitReached will be false when the amount of Zones created are less to the tier_limit
        expect(result.current.zoneLimitReached).toBe(false);
    });

    it('returns when label limit is reached', () => {
        const MOCK_CONFIG = {
            data: [
                {
                    key: 'analysis.tiering',
                    name: 'Multi-Tier Analysis Configuration',
                    description:
                        'This configuration parameter determines the limits of tiering with respect to analysis',
                    value: {
                        tier_limit: 1,
                        //How many labels can be created?
                        label_limit: 1,
                        multi_tier_analysis_enabled: true,
                    },
                    id: 8,
                    created_at: '2025-11-03T17:03:28.42299Z',
                    updated_at: '2025-11-03T17:03:28.42299Z',
                    deleted_at: {
                        Time: '0001-01-01T00:00:00Z',
                        Valid: false,
                    },
                },
            ],
        };

        const mockState = [
            {
                key: privilegeZonesKeys.tags(),
                //Create as many labels as needed to match the label_limit
                data: [createAssetGroupTag(1, undefined, AssetGroupTagTypeLabel)],
            },
            { key: configurationKeys.all, data: MOCK_CONFIG },
            { key: ConfigurationKey.Tiering, data: MOCK_CONFIG },
        ];
        const queryClient = setUpQueryClient(mockState);

        const { result } = renderHook(() => useTagLimits(), { queryClient });
        // labelLimitReached will be true when the amount of Labels created are equal to the label_limit
        expect(result.current.labelLimitReached).toBe(true);
    });

    it('does not return when label limit is not reached', () => {
        const MOCK_CONFIG = {
            data: [
                {
                    key: 'analysis.tiering',
                    name: 'Multi-Tier Analysis Configuration',
                    description:
                        'This configuration parameter determines the limits of tiering with respect to analysis',
                    value: {
                        tier_limit: 1,
                        //How many labels can be created?
                        label_limit: 3,
                        multi_tier_analysis_enabled: true,
                    },
                    id: 8,
                    created_at: '2025-11-03T17:03:28.42299Z',
                    updated_at: '2025-11-03T17:03:28.42299Z',
                    deleted_at: {
                        Time: '0001-01-01T00:00:00Z',
                        Valid: false,
                    },
                },
            ],
        };

        const mockState = [
            {
                key: privilegeZonesKeys.tags(),
                //Create as many labels as needed without reaching label_limit
                data: [
                    createAssetGroupTag(1, undefined, AssetGroupTagTypeLabel),
                    createAssetGroupTag(1, undefined, AssetGroupTagTypeLabel),
                ],
            },
            { key: configurationKeys.all, data: MOCK_CONFIG },
            { key: ConfigurationKey.Tiering, data: MOCK_CONFIG },
        ];
        const queryClient = setUpQueryClient(mockState);

        const { result } = renderHook(() => useTagLimits(), { queryClient });
        // labelLimitReached will be false when the amount of Labels created are less to the label_limit
        expect(result.current.labelLimitReached).toBe(false);
    });

    it('returns the remaining amount of zones available for creation', () => {
        const MOCK_CONFIG = {
            data: [
                {
                    key: 'analysis.tiering',
                    name: 'Multi-Tier Analysis Configuration',
                    description:
                        'This configuration parameter determines the limits of tiering with respect to analysis',
                    value: {
                        //How many zones can be created?
                        tier_limit: 10,
                        label_limit: 10,
                        multi_tier_analysis_enabled: true,
                    },
                    id: 8,
                    created_at: '2025-11-03T17:03:28.42299Z',
                    updated_at: '2025-11-03T17:03:28.42299Z',
                    deleted_at: {
                        Time: '0001-01-01T00:00:00Z',
                        Valid: false,
                    },
                },
            ],
        };

        const mockState = [
            {
                key: privilegeZonesKeys.tags(),
                //Create as many zones as needed but it should be less than tier_limit
                data: [createAssetGroupTag(), createAssetGroupTag()],
            },
            { key: configurationKeys.all, data: MOCK_CONFIG },
            { key: ConfigurationKey.Tiering, data: MOCK_CONFIG },
        ];
        const queryClient = setUpQueryClient(mockState);

        const { result } = renderHook(() => useTagLimits(), { queryClient });
        // remainingZonesAvailable will return the result of tier_limit - zones that have been created
        expect(result.current.remainingZonesAvailable).toBe(8);
    });

    it('returns the remaining amount of labels available for creation', () => {
        const MOCK_CONFIG = {
            data: [
                {
                    key: 'analysis.tiering',
                    name: 'Multi-Tier Analysis Configuration',
                    description:
                        'This configuration parameter determines the limits of tiering with respect to analysis',
                    value: {
                        tier_limit: 10,
                        //How many labels can be created?
                        label_limit: 10,
                        multi_tier_analysis_enabled: true,
                    },
                    id: 8,
                    created_at: '2025-11-03T17:03:28.42299Z',
                    updated_at: '2025-11-03T17:03:28.42299Z',
                    deleted_at: {
                        Time: '0001-01-01T00:00:00Z',
                        Valid: false,
                    },
                },
            ],
        };

        const mockState = [
            {
                key: privilegeZonesKeys.tags(),
                //Create as many labels as needed but it should be less than label_limit
                data: [
                    createAssetGroupTag(1, undefined, AssetGroupTagTypeLabel),
                    createAssetGroupTag(1, undefined, AssetGroupTagTypeLabel),
                ],
            },
            { key: configurationKeys.all, data: MOCK_CONFIG },
            { key: ConfigurationKey.Tiering, data: MOCK_CONFIG },
        ];
        const queryClient = setUpQueryClient(mockState);

        const { result } = renderHook(() => useTagLimits(), { queryClient });
        // labelLimitReached will return the result of label_limit - labels that have been created
        expect(result.current.remainingLabelsAvailable).toBe(8);
    });
});
