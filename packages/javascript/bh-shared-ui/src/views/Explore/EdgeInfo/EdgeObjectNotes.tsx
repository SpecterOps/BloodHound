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
import { Button, CircularProgress, Typography } from '@mui/material';
import { RedTeamNote } from 'js-client-library';
import React, { useState } from 'react';
import { NoteCard, NoteDetailDialog, NoteEditorDialog } from '../../../components/RedTeamNotes';
import { useDeleteRedTeamNote, useRedTeamNotesByEdgeKind } from '../../../hooks';
import { useNotifications } from '../../../providers';
import EdgeInfoCollapsibleSection from './EdgeInfoCollapsibleSection';

const sectionLabel = 'Red Team Notes';

interface EdgeObjectNotesProps {
    edgeKind: string;
}

// EdgeObjectNotes surfaces red team knowledge linked to the selected edge kind,
// keeping technique tradecraft and tooling notes next to the graph edge they
// belong to (e.g. DCSync tooling notes on a DCSync edge).
const EdgeObjectNotes: React.FC<EdgeObjectNotesProps> = ({ edgeKind }) => {
    const [isExpanded, setIsExpanded] = useState(false);
    const [editorOpen, setEditorOpen] = useState(false);
    const [noteToEdit, setNoteToEdit] = useState<RedTeamNote | null>(null);
    const [noteToView, setNoteToView] = useState<RedTeamNote | null>(null);

    const { data, isLoading } = useRedTeamNotesByEdgeKind(edgeKind);
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
        <EdgeInfoCollapsibleSection label={sectionLabel} isExpanded={isExpanded} onChange={setIsExpanded}>
            <div className='flex flex-col gap-2'>
                {isLoading && (
                    <div className='flex justify-center py-2'>
                        <CircularProgress size={20} />
                    </div>
                )}
                {!isLoading && notes.length === 0 && (
                    <Typography variant='body2' className='opacity-70'>
                        No notes linked to the {edgeKind} edge yet.
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
                    data-testid='edge-object-notes-add'>
                    Add Note
                </Button>
            </div>
            <NoteEditorDialog
                open={editorOpen}
                onClose={() => setEditorOpen(false)}
                note={noteToEdit}
                defaultValues={{ edge_kind: edgeKind, type: 'technique' }}
            />
            <NoteDetailDialog
                open={!!noteToView}
                note={noteToView}
                onClose={() => setNoteToView(null)}
                onEdit={handleEditNote}
                onDelete={handleDeleteNote}
            />
        </EdgeInfoCollapsibleSection>
    );
};

export default EdgeObjectNotes;
