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
import { faTrashCan } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AssetGroupTagSelector } from 'js-client-library';
import { FC } from 'react';

const DeleteSelectorButton: FC<{
    ruleId: string;
    ruleData: AssetGroupTagSelector | undefined;
    onClick: () => void;
}> = ({ ruleId, ruleData: selectorData, onClick }) => {
    if (ruleId === '') return null;

    if (selectorData === undefined) return null;

    if (selectorData.is_default) return null;

    return (
        <Button data-testid='privilege-zones_save_selector-form_delete-button' variant={'text'} onClick={onClick}>
            <span>
                <FontAwesomeIcon icon={faTrashCan} /> Delete Rule
            </span>
        </Button>
    );
};

export default DeleteSelectorButton;
