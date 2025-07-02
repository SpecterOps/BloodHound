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
import { useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { setParamsFactory } from '../../utils';
import { useHighestPrivilegeTagId } from '../useAssetGroupTags/useAssetGroupTags';

export type ZoneQueryParams = {
    assetGroupTagId: number | undefined;
};

const parseAssetGroupTagId = (assetGroupTagId: string | null, topTagId: number | undefined): number | undefined => {
    if (assetGroupTagId !== null) {
        return parseInt(assetGroupTagId);
    }

    return topTagId;
};

export const useZoneQueryParams = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const topTagId = useHighestPrivilegeTagId();

    const assetGroupTagId = parseAssetGroupTagId(searchParams.get('assetGroupTagId'), topTagId);

    const params = new URLSearchParams();
    if (typeof assetGroupTagId === 'number') params.append('asset_group_tag_id', assetGroupTagId.toString());

    return {
        assetGroupTagId,
        params,
        setZoneQueryParams: useCallback(
            (updatedParams: Partial<ZoneQueryParams>) =>
                setParamsFactory(setSearchParams, ['assetGroupTagId'])(updatedParams),
            [setSearchParams]
        ),
    };
};
