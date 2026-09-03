// Copyright 2026 Specter Ops, Inc.
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

import { faPlus } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, Typography } from '@mui/material';
import { NodeDetails, NodeDetailsWithInfo, RedTeamNote } from 'js-client-library';
import React, { useState } from 'react';
import { useDeleteRedTeamNote, useRedTeamNotesByObjectId } from '../../hooks';
import { useNotifications } from '../../providers';
import { NoteCard, NoteDetailDialog, NoteEditorDialog } from '../RedTeamNotes';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';

const sectionLabel = 'Red Team Notes';

interface EntityObjectNotesProps {
    selectedNode: NodeDetails | NodeDetailsWithInfo;
}

const EntityObjectNotes: React.FC<EntityObjectNotesProps> = ({ selectedNode }) => {
    const objectId = (selectedNode.properties?.objectid as string | undefined) ?? '';

    const [isExpanded, setIsExpanded] = useState(false);
    const [editorOpen, setEditorOpen] = useState(false);
    const [noteToEdit, setNoteToEdit] = useState<RedTeamNote | null>(null);
    const [noteToView, setNoteToView] = useState<RedTeamNote | null>(null);

    const { data, isLoading, isError, error } = useRedTeamNotesByObjectId(objectId);
    const deleteNote = useDeleteRedTeamNote();
    const { addNotification } = useNotifications();

    const notes = data?.data ?? [];

    const handleAddNote = () => {
        setNoteToEdit(null);
        setEditorOpen(true);
    };

    const handleEditNote = (note: RedTeamNote) => {
        setNoteToView(null);
        setNoteToEdit(note);
        setEditorOpen(true);
    };

    const handleDeleteNote = (note: RedTeamNote) => {
        deleteNote.mutate(note.id, {
            onSuccess: () => {
                addNotification('Note deleted.', 'redTeamNoteDeleted', { variant: 'success' });
                setNoteToView(null);
            },
            onError: () => addNotification('Failed to delete note.', 'redTeamNoteDeleteFailed', { variant: 'error' }),
        });
    };

    return (
        <EntityInfoCollapsibleSection
            label={sectionLabel}
            isExpanded={isExpanded}
            onChange={setIsExpanded}
            isLoading={isLoading}
            isError={isError}
            error={error}>
            <div className='flex flex-col gap-2'>
                {notes.length === 0 && (
                    <Typography variant='body2' className='opacity-70'>
                        No notes linked to this object yet.
                    </Typography>
                )}
                {notes.map((note) => (
                    <NoteCard
                        key={note.id}
                        note={note}
                        compact
                        onOpen={setNoteToView}
                        onEdit={handleEditNote}
                        onDelete={handleDeleteNote}
                    />
                ))}
                <Button
                    size='small'
                    startIcon={<FontAwesomeIcon icon={faPlus} />}
                    onClick={handleAddNote}
                    data-testid='entity-object-notes-add'>
                    Add Note
                </Button>
            </div>
            <NoteEditorDialog
                open={editorOpen}
                onClose={() => setEditorOpen(false)}
                note={noteToEdit}
                defaultValues={{ object_id: objectId }}
            />
            <NoteDetailDialog
                open={!!noteToView}
                note={noteToView}
                onClose={() => setNoteToView(null)}
                onEdit={handleEditNote}
                onDelete={handleDeleteNote}
            />
        </EntityInfoCollapsibleSection>
    );
};

export default EntityObjectNotes;
