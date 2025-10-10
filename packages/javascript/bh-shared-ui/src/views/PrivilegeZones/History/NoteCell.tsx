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
import { SystemString } from 'js-client-library';
import { AppIcon } from '../../../components';
import { useHistoryTableContext } from './HistoryTableContext';
import { HistoryItem } from './types';

export const NoteCell = ({ row }: { row: { original: Partial<HistoryItem> } }) => {
    const { currentNote, setCurrentNote } = useHistoryTableContext();
    const { email, note, date, actor } = row.original;

    const handleOnClick = () => {
        const selectedNote = {
            note: note,
            createdBy: email,
            timestamp: date,
        };

        // If the same note that is selected is clicked on, hide the note
        if (
            currentNote &&
            currentNote.note === note &&
            currentNote.createdBy === email &&
            currentNote.timestamp === date
        )
            setCurrentNote(null);
        else setCurrentNote(selectedNote);
    };

    return (
        <div className='w-full flex justify-center'>
            {actor === SystemString ? (
                <Tooltip tooltip={`No notes for ${SystemString} history`}>
                    <p>-</p>
                </Tooltip>
            ) : (
                <Tooltip tooltip={!note ? 'No notes' : currentNote ? 'Hide note' : 'Show note'}>
                    <span>
                        <Button
                            variant={'text'}
                            className='disabled:opacity-25'
                            onClick={handleOnClick}
                            disabled={!note}>
                            <AppIcon.LinedPaper size={24} />
                        </Button>
                    </span>
                </Tooltip>
            )}
        </div>
    );
};
