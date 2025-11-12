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
import { AssetGroupTag } from 'js-client-library';
import { useAssetGroupTags, useHighestPrivilegeTag } from './useAssetGroupTags';
import { usePZQueryParams } from './usePZParams';

export const HYGIENE_AGT_ID = 0;
export const HYGIENE_TAG_NAME = 'Hygiene';

export const HygieneTag = {
    name: HYGIENE_TAG_NAME,
    id: HYGIENE_AGT_ID,
    glyph: null,
} as const;

export const useSelectedTag = (): AssetGroupTag | typeof HygieneTag | undefined => {
    const tags = useAssetGroupTags().data ?? [];
    const { assetGroupTagId } = usePZQueryParams();
    const { tag: highestPrivilegeTag } = useHighestPrivilegeTag();

    if (assetGroupTagId === HYGIENE_AGT_ID) return HygieneTag;

    return tags.find((tag) => tag.id === assetGroupTagId) ?? highestPrivilegeTag;
};
