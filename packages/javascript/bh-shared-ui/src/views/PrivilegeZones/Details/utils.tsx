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

import { AssetGroupTag, AssetGroupTagSelector, SeedTypes } from 'js-client-library';

export const isTag = (data: any): data is AssetGroupTag => {
    return 'kind_id' in data;
};

export const isRule = (data: any): data is AssetGroupTagSelector => {
    return 'is_default' in data;
};

export const getRuleSeedType = (rule: AssetGroupTagSelector): SeedTypes => {
    const firstSeed = rule.seeds[0];

    return firstSeed.type;
};

export const getListHeight = (windoHeight: number) => {
    if (windoHeight > 1080) return 760;
    if (1080 >= windoHeight && windoHeight > 900) return 640;
    if (900 >= windoHeight) return 436;
    return 436;
};
