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

import { faBook, faDownload, faPlus, faSearch, faUpload } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Autocomplete, Button, CircularProgress, InputAdornment, MenuItem, TextField, Typography } from '@mui/material';
import {
    DeleteConfirmationDialog,
    NOTE_SORT_OPTIONS,
    NOTE_TYPE_DESCRIPTIONS,
    NOTE_TYPE_ICONS,
    NOTE_TYPE_OPTIONS,
    NoteCard,
    NoteDetailDialog,
    NoteEditorDialog,
    PageWithTitle,
    listRedTeamNotes,
    uploadRedTeamNoteAttachment,
    useCreateRedTeamNote,
    useDeleteRedTeamNote,
    useNotifications,
    useRedTeamNoteTags,
    useRedTeamNotes,
} from 'bh-shared-ui';
import { RedTeamNote, RedTeamNoteType, RedTeamNotesListParams } from 'js-client-library';
import React, { ChangeEvent, useEffect, useRef, useState } from 'react';
import { exportNotesZip, importNotesZip } from './zipUtils';

const PAGE_SIZE = 25;

const RedTeamNotes: React.FC = () => {
    const [searchInput, setSearchInput] = useState('');
    const [search, setSearch] = useState('');
    const [typeFilter, setTypeFilter] = useState<RedTeamNoteType | ''>('');
    const [tagFilter, setTagFilter] = useState<string[]>([]);
    const [sort, setSort] = useState('-updated_at');
    const [page, setPage] = useState(0);

    const [editorOpen, setEditorOpen] = useState(false);
    const [noteToEdit, setNoteToEdit] = useState<RedTeamNote | null>(null);
    const [noteToView, setNoteToView] = useState<RedTeamNote | null>(null);
    const [noteToDelete, setNoteToDelete] = useState<RedTeamNote | null>(null);
    const [isExporting, setIsExporting] = useState(false);
    const [isImporting, setIsImporting] = useState(false);

    const importInputRef = useRef<HTMLInputElement>(null);

    const { addNotification } = useNotifications();
    const createNote = useCreateRedTeamNote();

    useEffect(() => {
        const debounceHandle = setTimeout(() => {
            setSearch(searchInput);
            setPage(0);
        }, 300);

        return () => clearTimeout(debounceHandle);
    }, [searchInput]);

    const params: RedTeamNotesListParams = {
        search: search || undefined,
        type: typeFilter || undefined,
        tags: tagFilter.length > 0 ? tagFilter : undefined,
        sort,
        skip: page * PAGE_SIZE,
        limit: PAGE_SIZE,
    };

    const { data, isLoading, isError } = useRedTeamNotes(params);
    const deleteNote = useDeleteRedTeamNote();
    const { data: tagSuggestions } = useRedTeamNoteTags();

    const notes = data?.data ?? [];
    const totalCount = data?.count ?? 0;
    const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE));

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
        setNoteToView(null);
        setNoteToDelete(note);
    };

    const handleConfirmDelete = () => {
        if (!noteToDelete) return;

        deleteNote.mutate(noteToDelete.id, {
            onSuccess: () => {
                addNotification('Note deleted.', 'redTeamNoteDeleted', { variant: 'success' });
                setNoteToDelete(null);
            },
            onError: () => {
                addNotification('Failed to delete note.', 'redTeamNoteDeleteFailed', { variant: 'error' });
                setNoteToDelete(null);
            },
        });
    };

    const handleExport = async () => {
        try {
            setIsExporting(true);
            const { data: allNotes } = await listRedTeamNotes({ limit: 500 });

            const { blob, stats } = await exportNotesZip(allNotes, (token) =>
                fetch(`/api/v2/red-team-notes/media/${token}`)
            );

            const timestamp = new Date().toISOString().slice(0, 10);
            const objectUrl = URL.createObjectURL(blob);
            const anchor = document.createElement('a');
            anchor.href = objectUrl;
            anchor.download = `red-team-notes-${timestamp}.zip`;
            anchor.click();
            URL.revokeObjectURL(objectUrl);

            addNotification(
                `Exported ${stats.noteCount} notes (${stats.attachmentCount} attachments) to ZIP.`,
                'redTeamNotesExported',
                { variant: 'success' }
            );
        } catch {
            addNotification('Export failed.', 'redTeamNotesExportFailed', { variant: 'error' });
        } finally {
            setIsExporting(false);
        }
    };

    const handleImportFileSelected = async (event: ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        event.target.value = '';

        if (!file) return;

        try {
            setIsImporting(true);
            const result = await importNotesZip(
                file,
                (attachmentFile) => uploadRedTeamNoteAttachment(attachmentFile),
                (payload) => createNote.mutateAsync(payload)
            );
            addNotification(
                `Imported ${result.createdNotes} notes (${result.createdAttachments} attachments).`,
                'redTeamNotesImported',
                { variant: 'success' }
            );
        } catch {
            addNotification('Import failed. Expected a ZIP with markdown notes.', 'redTeamNotesImportFailed', {
                variant: 'error',
            });
        } finally {
            setIsImporting(false);
        }
    };

    const isFiltering = !!search || !!typeFilter || tagFilter.length > 0;

    return (
        <PageWithTitle
            title='Red Team Knowledge Base'
            fullWidth
            pageDescription={
                <Typography variant='body2' className='opacity-70'>
                    Team tradecraft, offensive tooling references and source material — linked to the graph.
                </Typography>
            }>
            <div className='flex flex-col gap-4 pb-8'>
                <div className='flex items-center gap-4 flex-wrap'>
                    <TextField
                        size='small'
                        placeholder='Search title or content'
                        value={searchInput}
                        onChange={(event) => setSearchInput(event.target.value)}
                        className='w-80'
                        InputProps={{
                            startAdornment: (
                                <InputAdornment position='start'>
                                    <FontAwesomeIcon icon={faSearch} />
                                </InputAdornment>
                            ),
                        }}
                        data-testid='red-team-notes-search'
                    />
                    <TextField
                        select
                        size='small'
                        label='Type'
                        value={typeFilter}
                        onChange={(event) => {
                            setTypeFilter(event.target.value as RedTeamNoteType | '');
                            setPage(0);
                        }}
                        className='w-44'
                        data-testid='red-team-notes-type-filter'>
                        <MenuItem value=''>All types</MenuItem>
                        {NOTE_TYPE_OPTIONS.map((option) => (
                            <MenuItem key={option.value} value={option.value}>
                                <span className='flex items-center gap-2'>
                                    <FontAwesomeIcon icon={NOTE_TYPE_ICONS[option.value]} fixedWidth />
                                    {option.label}
                                </span>
                            </MenuItem>
                        ))}
                    </TextField>
                    <Autocomplete
                        multiple
                        freeSolo
                        size='small'
                        className='w-64'
                        options={(tagSuggestions ?? []).map((tagCount) => tagCount.tag)}
                        value={tagFilter}
                        onChange={(_event, newValue) => {
                            setTagFilter(newValue as string[]);
                            setPage(0);
                        }}
                        renderInput={(inputParams) => (
                            <TextField {...inputParams} label='Tags' placeholder='Filter by tags' />
                        )}
                        data-testid='red-team-notes-tag-autocomplete'
                    />
                    <TextField
                        select
                        size='small'
                        label='Sort'
                        value={sort}
                        onChange={(event) => {
                            setSort(event.target.value);
                            setPage(0);
                        }}
                        className='w-48'
                        data-testid='red-team-notes-sort'>
                        {NOTE_SORT_OPTIONS.map((option) => (
                            <MenuItem key={option.value} value={option.value}>
                                {option.label}
                            </MenuItem>
                        ))}
                    </TextField>
                    <div className='ml-auto flex items-center gap-2'>
                        <input
                            type='file'
                            accept='.zip,application/zip'
                            hidden
                            ref={importInputRef}
                            onChange={handleImportFileSelected}
                            data-testid='red-team-notes-import-input'
                        />
                        <Button
                            startIcon={<FontAwesomeIcon icon={faUpload} />}
                            onClick={() => importInputRef.current?.click()}
                            disabled={isImporting}
                            data-testid='red-team-notes-import'>
                            {isImporting ? 'Importing...' : 'Import ZIP'}
                        </Button>
                        <Button
                            startIcon={
                                isExporting ? <CircularProgress size={14} /> : <FontAwesomeIcon icon={faDownload} />
                            }
                            onClick={handleExport}
                            disabled={notes.length === 0 || isExporting}
                            data-testid='red-team-notes-export'>
                            Export ZIP
                        </Button>
                        <Button
                            variant='contained'
                            startIcon={<FontAwesomeIcon icon={faPlus} />}
                            onClick={handleAddNote}
                            data-testid='red-team-notes-create'>
                            New Note
                        </Button>
                    </div>
                </div>

                {isLoading && (
                    <div className='flex justify-center py-8'>
                        <CircularProgress />
                    </div>
                )}

                {isError && (
                    <Typography color='error' data-testid='red-team-notes-error'>
                        Failed to load notes.
                    </Typography>
                )}

                {!isLoading && !isError && notes.length === 0 && (
                    <div className='flex flex-col items-center gap-3 py-16' data-testid='red-team-notes-empty'>
                        <FontAwesomeIcon icon={faBook} size='3x' className='opacity-40' />
                        <Typography variant='h6' className='opacity-80'>
                            {isFiltering ? 'No notes match your filters.' : 'Your knowledge base is empty.'}
                        </Typography>
                        {!isFiltering && (
                            <div className='flex flex-col gap-1 items-center opacity-60 text-sm'>
                                <span>Capture techniques, tool usage and source references for the whole team.</span>
                                <span>
                                    Notes can be linked to graph objects and edge kinds — they show up in the Explore
                                    panel.
                                </span>
                            </div>
                        )}
                        {!isFiltering && (
                            <Button
                                variant='contained'
                                startIcon={<FontAwesomeIcon icon={faPlus} />}
                                onClick={handleAddNote}
                                className='mt-2'>
                                Create your first note
                            </Button>
                        )}
                    </div>
                )}

                {!isLoading && !isError && notes.length > 0 && (
                    <>
                        <Typography variant='body2' className='opacity-70'>
                            {totalCount} {totalCount === 1 ? 'note' : 'notes'}
                            {typeFilter && ` — ${NOTE_TYPE_DESCRIPTIONS[typeFilter]}`}
                        </Typography>
                        <div className='flex flex-col gap-3'>
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
                        </div>
                    </>
                )}

                {totalCount > PAGE_SIZE && (
                    <div className='flex items-center justify-end gap-4'>
                        <Typography variant='body2' className='opacity-70'>
                            Page {page + 1} of {totalPages}
                        </Typography>
                        <Button size='small' disabled={page === 0} onClick={() => setPage(page - 1)}>
                            Previous
                        </Button>
                        <Button size='small' disabled={page + 1 >= totalPages} onClick={() => setPage(page + 1)}>
                            Next
                        </Button>
                    </div>
                )}
            </div>

            <NoteEditorDialog open={editorOpen} onClose={() => setEditorOpen(false)} note={noteToEdit} />

            <NoteDetailDialog
                open={!!noteToView}
                note={noteToView}
                onClose={() => setNoteToView(null)}
                onEdit={handleEditNote}
                onDelete={handleDeleteNote}
            />

            <DeleteConfirmationDialog
                open={!!noteToDelete}
                itemName={noteToDelete?.title ?? ''}
                itemType='note'
                onCancel={() => setNoteToDelete(null)}
                onConfirm={handleConfirmDelete}
                isLoading={deleteNote.isLoading}
            />
        </PageWithTitle>
    );
};

export default RedTeamNotes;
