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

import { createAssetGroupTag } from '../../mocks/factories/privilegeZones';
import { renderHook, waitFor } from '../../test-utils';
import { setUpQueryClient } from '../../utils';
import { privilegeZonesKeys } from '../useAssetGroupTags';
import { configurationKeys } from '../useConfiguration';
import { featureFlagKeys } from '../useFeatureFlags';
import { useTagLimits } from './useTagLimits';

describe('useTagLimits', () => {
    it('zones reaches limit when count is equal to limit'),
        async () => {
            const mockState = [
                {
                    key: featureFlagKeys.getKey(['tier_management_engine']),
                    data: [
                        {
                            id: 1,
                            key: 'Test',
                            name: 'BloodHound',
                            description: 'Test',
                            enabled: true,
                            user_updatable: false,
                        },
                    ],
                },
                {
                    key: privilegeZonesKeys.tags(),
                    data: [createAssetGroupTag()],
                },
                { key: configurationKeys.all, data: null },
            ];
            const queryClient = setUpQueryClient(mockState);

            const { result } = renderHook(() => useTagLimits(), { queryClient });

            await waitFor(() => expect(result.current.zoneLimitReached).toBe(false));
        };
});
