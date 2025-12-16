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

import { AssetGroupTagTypeLabel, AssetGroupTagTypeZone, parseTieringConfiguration } from 'js-client-library';
import { useAssetGroupTags } from '../useAssetGroupTags';
import { useGetConfiguration } from '../useConfiguration';

export const useTagLimits = () => {
    const tagsQuery = useAssetGroupTags();
    const { data } = useGetConfiguration();
    const config = parseTieringConfiguration(data);

    if (tagsQuery.isLoading || tagsQuery.isError || !config || !tagsQuery.isSuccess) {
        return {
            zoneLimitReached: true,
            labelLimitReached: true,
            remainingZonesAvailable: 0,
            remainingLabelsAvailable: 0,
        };
    }

    let zonesCount = 0;
    let labelsCount = 0;

    tagsQuery.data.forEach((tag) => {
        if (tag.type === AssetGroupTagTypeZone) {
            zonesCount += 1;
        }

        if (tag.type === AssetGroupTagTypeLabel) {
            labelsCount += 1;
        }
    });

    const { tier_limit, label_limit } = config.value;
    const zoneLimitReached = zonesCount >= tier_limit;
    const labelLimitReached = labelsCount >= label_limit;
    const remainingZonesAvailable = tier_limit - zonesCount;
    const remainingLabelsAvailable = label_limit - labelsCount;

    return { zoneLimitReached, labelLimitReached, remainingZonesAvailable, remainingLabelsAvailable };
};
