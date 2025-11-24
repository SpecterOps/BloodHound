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

import { Button } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import { AppLink } from '../../../components/Navigation';
import { usePZPathParams } from '../../../hooks';
import { labelsPath, privilegeZonesPath, rulesPath, savePath, zonesPath } from '../../../routes';

enum TitleSuffix {
    Selector = 'Rule',
    Zone = 'Zone',
    Label = 'Label',
}

export const getSavePath = (
    zoneId: string | undefined,
    labelId: string | undefined,
    selectorId: string | undefined
) => {
    const tagType = !labelId ? zonesPath : labelsPath;
    const tagPathId = tagType === 'zones' ? zoneId ?? '' : labelId ?? '';

    if (tagPathId === '') return;

    const dynamicSavePath = selectorId ? `${rulesPath}/${selectorId}/${savePath}` : savePath;

    return `/${privilegeZonesPath}/${tagType}/${tagPathId}/${dynamicSavePath}`;
};

export const suffix = (zoneId: string | undefined, labelId: string | undefined, selectorId: string | undefined) => {
    if (selectorId) {
        return TitleSuffix.Selector;
    } else if (labelId) {
        return TitleSuffix.Label;
    } else {
        return TitleSuffix.Zone;
    }
};

export const PZEditButton: FC<{
    showEditButton: boolean;
}> = ({ showEditButton }) => {
    const { zoneId, labelId, selectorId } = usePZPathParams();
    const saveLink = getSavePath(zoneId, labelId, selectorId);
    const titleSuffix = suffix(zoneId, labelId, selectorId);

    return (
        <div className='flex flex-col gap-4 w-[6.75rem]'>
            {showEditButton && (
                <Button asChild={!!saveLink} variant={'secondary'} disabled={!saveLink}>
                    <AppLink data-testid='privilege-zones_edit-button' to={saveLink || ''}>
                        Edit {titleSuffix}
                    </AppLink>
                </Button>
            )}
        </div>
    );
};
