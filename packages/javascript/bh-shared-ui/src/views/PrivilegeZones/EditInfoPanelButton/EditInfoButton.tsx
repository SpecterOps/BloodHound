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
enum TitleAppend {
    Selector = 'Rule',
    Zone = 'Zone',
    Label = 'Label',
}

const EditInfoButton: FC<{
    labelId?: string | undefined;
    selectorId?: string | undefined;
    zoneId?: string | undefined;
    showEditButton: boolean;
    saveLink?: string | undefined;
}> = ({ labelId, selectorId, zoneId, showEditButton, saveLink }) => {
    let titleAppend: TitleAppend | undefined;

    if (selectorId) {
        titleAppend = TitleAppend.Selector;
    } else if (zoneId) {
        titleAppend = TitleAppend.Zone;
    } else if (labelId) {
        titleAppend = TitleAppend.Label;
    }

    return (
        <div className='flex flex-col gap-4 basis-1/3'>
            <Button asChild={showEditButton || !saveLink} variant={'secondary'} disabled={!!saveLink}>
                <AppLink className='w-[6.75rem]' data-testid='privilege-zones_edit-button' to={saveLink || ''}>
                    Edit {titleAppend}
                </AppLink>
            </Button>
        </div>
    );
};

export default EditInfoButton;
