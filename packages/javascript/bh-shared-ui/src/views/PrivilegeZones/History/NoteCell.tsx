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

import { Button, Tooltip } from '@bloodhoundenterprise/doodleui';
import { BloodHoundString } from 'js-client-library';
import { AppIcon } from '../../../components/AppIcon';
import { useHistoryTableContext } from './HistoryTableContext';
import { HistoryItem } from './types';

export const NoteCell = ({ row }: { row: { original: HistoryItem } }) => {
    const { selected, setSelected, clearSelected } = useHistoryTableContext();
    const { note, id, actor } = row.original;

    const noteIsActive = selected?.id === id;

    const handleClick = () => (noteIsActive ? clearSelected() : setSelected(row.original));

    return (
        <div className='w-full flex justify-center'>
            {actor === BloodHoundString ? (
                <Tooltip tooltip={`No notes for ${BloodHoundString} history`}>
                    <p>-</p>
                </Tooltip>
            ) : (
                <Tooltip tooltip={!note ? 'No notes' : noteIsActive ? 'Hide note' : 'Show note'}>
                    <span>
                        <Button variant={'text'} className='disabled:opacity-25' onClick={handleClick} disabled={!note}>
                            <AppIcon.LinedPaper size={24} className='-mb-[3px]' />
                        </Button>
                    </span>
                </Tooltip>
            )}
        </div>
    );
};
