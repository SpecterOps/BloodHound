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

import { type AssetGroupTag } from 'js-client-library';
import { isNode, ItemResponse } from '../hooks';
import { detailsPath, labelsPath, privilegeZonesPath, zonesPath } from '../routes';

export function getOwnedObjectAssetGroupTagRemoveNodePath(assetGroupTag?: AssetGroupTag) {
    return `/${privilegeZonesPath}/${labelsPath}/${assetGroupTag?.id}/${detailsPath}`;
}

export function getTierZeroAssetGroupTagRemoveNodePath(assetGroupTag?: AssetGroupTag) {
    return `/${privilegeZonesPath}/${zonesPath}/${assetGroupTag?.id}/${detailsPath}`;
}

export function isOwnedObject(item: ItemResponse): boolean {
    if (!isNode(item)) return false;

    return item.isOwnedObject;
}

export function isTierZero(item: ItemResponse): boolean {
    if (!isNode(item)) return false;

    return item.isTierZero;
}
