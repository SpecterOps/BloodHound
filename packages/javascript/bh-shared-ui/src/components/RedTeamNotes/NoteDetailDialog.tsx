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

import { faCodeBranch, faCopy, faLink, faPen, faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, Chip, Dialog, DialogActions, DialogContent, DialogTitle, Divider, IconButton, Tooltip } from '@mui/material';
import { RedTeamNote } from 'js-client-library';
import React from 'react';
import { copyToClipboard } from '../../utils';
import MarkdownContent from '../MarkdownContent';
import { NOTE_TYPE_COLORS, NOTE_TYPE_ICONS, NOTE_TYPE_LABELS } from './constants';
import { formatAbsoluteTime } from './utils';

export interface NoteDetailDialogProps {
    open: boolean;
    note: RedTeamNote | null;
    onClose: () => void;
    onEdit?: (note: RedTeamNote) => void;
    onDelete?: (note: RedTeamNote) => void;
}

const NoteDetailDialog: React.FC<NoteDetailDialogProps> = ({ open, note, onClose, onEdit, onDelete }) => {
    if (!note) return null;

    return (
        <Dialog
            open={open}
            onClose={onClose}
            fullWidth
            maxWidth='md'
            PaperProps={{ sx: { maxHeight: '86vh', display: 'flex', flexDirection: 'column' } }}
            data-testid='red-team-note-detail-dialog'>
            <DialogTitle>
                <div className='flex items-center gap-3 flex-wrap'>
                    <FontAwesomeIcon icon={NOTE_TYPE_ICONS[note.type] ?? NOTE_TYPE_ICONS.general} />
                    <span>{note.title}</span>
                    <Chip
                        label={NOTE_TYPE_LABELS[note.type] ?? note.type}
                        color={NOTE_TYPE_COLORS[note.type] ?? 'default'}
                        size='small'
                        variant='outlined'
                    />
                    {note.tags.map((tag) => (
                        <Chip key={tag} label={tag} size='small' />
                    ))}
                </div>
            </DialogTitle>
            <DialogContent className='flex flex-col gap-3' sx={{ overflowY: 'auto', minHeight: 0 }}>
                {note.url && (
                    <a href={note.url} target='_blank' rel='noreferrer' className='flex items-center gap-2 text-sm'>
                        <FontAwesomeIcon icon={faLink} /> {note.url}
                    </a>
                )}
                <Divider />
                {note.content ? (
                    <div className='text-sm' data-testid='red-team-note-detail-content'>
                        <MarkdownContent markdown={note.content} />
                    </div>
                ) : (
                    <span className='opacity-70 text-sm'>This note has no content yet.</span>
                )}
                <Divider />
                <div className='flex items-center gap-4 text-xs opacity-70 flex-wrap'>
                    {note.object_id && (
                        <span className='flex items-center gap-1'>
                            <FontAwesomeIcon icon={faCodeBranch} /> object: {note.object_id}
                        </span>
                    )}
                    {note.edge_kind && <span>edge: {note.edge_kind}</span>}
                    <span>created {formatAbsoluteTime(note.created_at)}</span>
                    <span>updated {formatAbsoluteTime(note.updated_at)}</span>
                </div>
            </DialogContent>
            <DialogActions>
                <Tooltip title='Copy content'>
                    <IconButton
                        onClick={() => copyToClipboard(note.content)}
                        className='mr-auto ml-2'
                        data-testid='red-team-note-detail-copy'>
                        <FontAwesomeIcon icon={faCopy} />
                    </IconButton>
                </Tooltip>
                {onDelete && (
                    <Button
                        color='error'
                        startIcon={<FontAwesomeIcon icon={faTrash} />}
                        onClick={() => onDelete(note)}
                        data-testid='red-team-note-detail-delete'>
                        Delete
                    </Button>
                )}
                {onEdit && (
                    <Button
                        startIcon={<FontAwesomeIcon icon={faPen} />}
                        onClick={() => onEdit(note)}
                        data-testid='red-team-note-detail-edit'>
                        Edit
                    </Button>
                )}
                <Button onClick={onClose} variant='contained' data-testid='red-team-note-detail-close'>
                    Close
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default NoteDetailDialog;
