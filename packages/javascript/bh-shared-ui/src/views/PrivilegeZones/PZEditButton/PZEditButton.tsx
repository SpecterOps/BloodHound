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
    Rule = 'Rule',
    Zone = 'Zone',
    Label = 'Label',
}

export const getSavePath = (zoneId: string | undefined, labelId: string | undefined) => {
    const tagType = !labelId ? zonesPath : labelsPath;
    const tagPathId = tagType === 'zones' ? zoneId ?? '' : labelId ?? '';

    if (tagPathId === '') return;

    return `/${privilegeZonesPath}/${tagType}/${tagPathId}/${savePath}`;
};

export const getRuleLink = (zoneId: string | undefined, labelId: string | undefined, ruleId: string | undefined) => {
    const tagType = !labelId ? zonesPath : labelsPath;
    const tagPathId = tagType === 'zones' ? zoneId ?? '' : labelId ?? '';

    if (tagPathId === '') return;

    return `/${privilegeZonesPath}/${tagType}/${tagPathId}/${rulesPath}/${ruleId}/${savePath}`;
};

export const suffix = (zoneId: string | undefined, labelId: string | undefined) => {
    if (labelId) {
        return TitleSuffix.Label;
    } else {
        return TitleSuffix.Zone;
    }
};

export const PZEditButton: FC = () => {
    const { zoneId, labelId, ruleId } = usePZPathParams();
    const saveLink = getSavePath(zoneId, labelId);
    const ruleLink = getRuleLink(zoneId, labelId, ruleId);
    const titleSuffix = suffix(zoneId, labelId);

    return (
        <div className='flex  gap-4 w-[6.75rem]'>
            <Button asChild={!!saveLink} variant={'secondary'} disabled={!saveLink}>
                <AppLink data-testid='privilege-zones_edit-button' to={saveLink || ''}>
                    Edit {titleSuffix}
                </AppLink>
            </Button>
            <Button variant={'secondary'} disabled={!ruleId}>
                <AppLink data-testid='privilege-zones_edit-button' to={ruleLink || ''}>
                    Edit Rule
                </AppLink>
            </Button>
        </div>
    );
};
