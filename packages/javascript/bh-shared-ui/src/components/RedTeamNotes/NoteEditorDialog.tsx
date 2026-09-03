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

import { faEye, faImage, faKeyboard } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, MenuItem, Tab, Tabs, TextField } from '@mui/material';
import { RedTeamNote, RedTeamNotePayload, RedTeamNoteType } from 'js-client-library';
import React, { ChangeEvent, ClipboardEvent, useEffect, useRef, useState } from 'react';
import { uploadRedTeamNoteAttachment, useCreateRedTeamNote, useUpdateRedTeamNote } from '../../hooks';
import { useNotifications } from '../../providers';
import MarkdownContent from '../MarkdownContent';
import { NOTE_TYPE_OPTIONS } from './constants';

export interface NoteEditorDialogProps {
    open: boolean;
    onClose: () => void;
    note?: RedTeamNote | null;
    defaultValues?: Partial<RedTeamNotePayload>;
}

const emptyPayload: RedTeamNotePayload = {
    title: '',
    content: '',
    type: 'general',
    tags: [],
    url: '',
    object_id: '',
    edge_kind: '',
};

const NoteEditorDialog: React.FC<NoteEditorDialogProps> = ({ open, onClose, note, defaultValues }) => {
    const [payload, setPayload] = useState<RedTeamNotePayload>(emptyPayload);
    const [tagsInput, setTagsInput] = useState<string>('');
    const [titleError, setTitleError] = useState<string>('');
    const [contentTab, setContentTab] = useState<'write' | 'preview'>('write');
    const [isUploadingImage, setIsUploadingImage] = useState(false);

    const fileInputRef = useRef<HTMLInputElement>(null);
    const contentInputRef = useRef<HTMLTextAreaElement>(null);

    const createNote = useCreateRedTeamNote();
    const updateNote = useUpdateRedTeamNote();
    const { addNotification } = useNotifications();

    const insertMarkdownAtCursor = (markdownToInsert: string) => {
        const element = contentInputRef.current;
        const currentContent = payload.content;

        if (!element) {
            setPayload({
                ...payload,
                content: currentContent ? `${currentContent}\n\n${markdownToInsert}` : markdownToInsert,
            });
            return;
        }

        const cursorStart = element.selectionStart ?? currentContent.length;
        const cursorEnd = element.selectionEnd ?? cursorStart;
        const before = currentContent.slice(0, cursorStart);
        const after = currentContent.slice(cursorEnd);
        const linePrefix = before && !before.endsWith('\n') ? '\n' : '';
        const insertedText = `${linePrefix}${markdownToInsert}`;
        const nextContent = `${before}${insertedText}${after}`;
        const nextCursor = (before + insertedText).length;

        setPayload({ ...payload, content: nextContent });

        requestAnimationFrame(() => {
            element.focus();
            element.setSelectionRange(nextCursor, nextCursor);
        });
    };

    const uploadAndInsert = async (file: File) => {
        setIsUploadingImage(true);

        try {
            const attachment = await uploadRedTeamNoteAttachment(file);
            insertMarkdownAtCursor(attachment.markdown);
            addNotification('Image uploaded — markdown reference inserted.', 'redTeamImageUploaded', {
                variant: 'success',
            });
        } catch {
            addNotification(
                'Image upload failed. Only png, jpeg, gif, webp or svg under 5MB are allowed.',
                'redTeamImageUploadFailed',
                { variant: 'error' }
            );
        } finally {
            setIsUploadingImage(false);
        }
    };

    const handleImagePaste = (event: ClipboardEvent) => {
        const pastedImages = Array.from(event.clipboardData.files).filter((file) =>
            file.type.startsWith('image/')
        );

        if (pastedImages.length === 0) return;

        event.preventDefault();
        pastedImages.forEach((file) => uploadAndInsert(file));
    };

    const handleImageSelected = (event: ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        event.target.value = '';

        if (!file) return;

        uploadAndInsert(file);
    };

    useEffect(() => {
        if (!open) return;

        if (note) {
            setPayload({
                title: note.title,
                content: note.content,
                type: note.type,
                tags: note.tags,
                url: note.url,
                object_id: note.object_id,
                edge_kind: note.edge_kind,
            });
            setTagsInput(note.tags.join(', '));
        } else {
            const nextPayload = { ...emptyPayload, ...(defaultValues ?? {}) };
            setPayload(nextPayload);
            setTagsInput((nextPayload.tags ?? []).join(', '));
        }

        setTitleError('');
        setContentTab('write');
        createNote.reset();
        updateNote.reset();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [open, note, defaultValues]);

    const handleSave = () => {
        if (!payload.title.trim()) {
            setTitleError('Title is required');
            return;
        }

        const parsedTags = tagsInput
            .split(',')
            .map((tag) => tag.trim())
            .filter((tag) => tag.length > 0);

        const finalPayload: RedTeamNotePayload = { ...payload, tags: parsedTags };

        const onSuccess = () => {
            addNotification(note ? 'Note updated.' : 'Note created.', 'redTeamNoteSaved', { variant: 'success' });
            onClose();
        };

        const onError = () => {
            addNotification('Failed to save note. Please try again.', 'redTeamNoteSaveFailed', { variant: 'error' });
        };

        if (note) {
            updateNote.mutate({ noteId: note.id, payload: finalPayload }, { onSuccess, onError });
        } else {
            createNote.mutate(finalPayload, { onSuccess, onError });
        }
    };

    const isSaving = createNote.isLoading || updateNote.isLoading;

    return (
        <Dialog open={open} onClose={onClose} fullWidth maxWidth='md' data-testid='red-team-note-editor-dialog'>
            <DialogTitle>{note ? 'Edit Note' : 'New Red Team Note'}</DialogTitle>
            <DialogContent className='flex flex-col gap-4 pt-4' sx={{ overflowY: 'auto', maxHeight: '72vh' }}>
                <TextField
                    label='Title'
                    value={payload.title}
                    onChange={(event) => {
                        setPayload({ ...payload, title: event.target.value });
                        setTitleError('');
                    }}
                    error={!!titleError}
                    helperText={titleError}
                    fullWidth
                    size='small'
                    autoFocus
                    data-testid='red-team-note-editor-title'
                />
                <div className='flex gap-4'>
                    <TextField
                        select
                        label='Type'
                        value={payload.type}
                        onChange={(event) => setPayload({ ...payload, type: event.target.value as RedTeamNoteType })}
                        size='small'
                        className='w-48'
                        data-testid='red-team-note-editor-type'>
                        {NOTE_TYPE_OPTIONS.map((option) => (
                            <MenuItem key={option.value} value={option.value}>
                                {option.label}
                            </MenuItem>
                        ))}
                    </TextField>
                    <TextField
                        label='Tags (comma separated)'
                        value={tagsInput}
                        onChange={(event) => setTagsInput(event.target.value)}
                        size='small'
                        fullWidth
                        data-testid='red-team-note-editor-tags'
                    />
                </div>
                <TextField
                    label='URL (tool repo, reference, source code link)'
                    value={payload.url}
                    onChange={(event) => setPayload({ ...payload, url: event.target.value })}
                    size='small'
                    fullWidth
                    data-testid='red-team-note-editor-url'
                />
                <div className='flex gap-4'>
                    <TextField
                        label='Linked Object ID (optional)'
                        value={payload.object_id}
                        onChange={(event) => setPayload({ ...payload, object_id: event.target.value })}
                        size='small'
                        fullWidth
                        data-testid='red-team-note-editor-object-id'
                    />
                    <TextField
                        label='Linked Edge Kind (optional)'
                        value={payload.edge_kind}
                        onChange={(event) => setPayload({ ...payload, edge_kind: event.target.value })}
                        size='small'
                        fullWidth
                        data-testid='red-team-note-editor-edge-kind'
                    />
                </div>

                <div className='border border-neutral-3 rounded-lg overflow-hidden'>
                    <div className='flex items-center border-b border-neutral-3'>
                        <Tabs
                            value={contentTab}
                            onChange={(_event, value) => setContentTab(value)}
                            className='min-h-0 flex-1'>
                            <Tab
                                value='write'
                                className='min-h-0 py-2 text-xs'
                                icon={<FontAwesomeIcon icon={faKeyboard} />}
                                iconPosition='start'
                                label='Write'
                                data-testid='red-team-note-editor-tab-write'
                            />
                            <Tab
                                value='preview'
                                className='min-h-0 py-2 text-xs'
                                icon={<FontAwesomeIcon icon={faEye} />}
                                iconPosition='start'
                                label='Preview'
                                data-testid='red-team-note-editor-tab-preview'
                            />
                        </Tabs>
                        <input
                            type='file'
                            accept='image/png,image/jpeg,image/gif,image/webp,image/svg+xml'
                            hidden
                            ref={fileInputRef}
                            onChange={handleImageSelected}
                            data-testid='red-team-note-editor-image-input'
                        />
                        <Button
                            size='small'
                            className='mr-2'
                            disabled={isUploadingImage}
                            startIcon={
                                isUploadingImage ? <CircularProgress size={14} /> : <FontAwesomeIcon icon={faImage} />
                            }
                            onClick={() => fileInputRef.current?.click()}
                            data-testid='red-team-note-editor-insert-image'>
                            Insert image
                        </Button>
                    </div>
                    {contentTab === 'write' ? (
                        <TextField
                            placeholder={
                                'Markdown supported: # headings, **bold**, `code`, ```blocks```, lists, links...\nPaste an image (Ctrl+V) to upload it and insert the markdown reference automatically.'
                            }
                            value={payload.content}
                            onChange={(event) => setPayload({ ...payload, content: event.target.value })}
                            onPaste={handleImagePaste}
                            inputRef={contentInputRef}
                            fullWidth
                            multiline
                            minRows={10}
                            maxRows={14}
                            variant='outlined'
                            className='[&_.MuiOutlinedInput-notchedOutline]:border-none'
                            data-testid='red-team-note-editor-content'
                        />
                    ) : (
                        <div className='p-4 min-h-[230px] text-sm overflow-y-auto' data-testid='red-team-note-editor-preview'>
                            {payload.content ? (
                                <MarkdownContent markdown={payload.content} />
                            ) : (
                                <span className='opacity-70'>Nothing to preview yet.</span>
                            )}
                        </div>
                    )}
                </div>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose} color='inherit' data-testid='red-team-note-editor-cancel'>
                    Cancel
                </Button>
                <Button onClick={handleSave} variant='contained' disabled={isSaving} data-testid='red-team-note-editor-save'>
                    {isSaving ? 'Saving...' : 'Save'}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default NoteEditorDialog;
