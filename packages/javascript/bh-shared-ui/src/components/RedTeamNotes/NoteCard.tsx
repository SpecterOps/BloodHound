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

import { faClock, faCodeBranch, faCopy, faLink, faPen, faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Chip, IconButton, Tooltip, Typography } from '@mui/material';
import { RedTeamNote } from 'js-client-library';
import React from 'react';
import { adaptClickHandlerToKeyDown } from '../../utils/adaptClickHandlerToKeyDown';
import { copyToClipboard } from '../../utils';
import MarkdownContent from '../MarkdownContent';
import { NOTE_TYPE_COLORS, NOTE_TYPE_ICONS, NOTE_TYPE_LABELS } from './constants';
import { formatAbsoluteTime, formatRelativeTime, markdownExcerpt } from './utils';

export interface NoteCardProps {
    note: RedTeamNote;
    onOpen?: (note: RedTeamNote) => void;
    onEdit?: (note: RedTeamNote) => void;
    onDelete?: (note: RedTeamNote) => void;
    compact?: boolean;
}

const NOTE_CARD_ACCENTS: Record<string, string> = {
    general: 'border-l-neutral-3',
    technique: 'border-l-red-1',
    tool: 'border-l-blue-1',
    source: 'border-l-green-1',
};

const NOTE_CARD_ICON_COLORS: Record<string, string> = {
    general: 'text-neutral-3',
    technique: 'text-red-1',
    tool: 'text-blue-1',
    source: 'text-green-1',
};

const NoteCard: React.FC<NoteCardProps> = ({ note, onOpen, onEdit, onDelete, compact = false }) => {
    const accent = NOTE_CARD_ACCENTS[note.type] ?? NOTE_CARD_ACCENTS.general;
    const iconColor = NOTE_CARD_ICON_COLORS[note.type] ?? NOTE_CARD_ICON_COLORS.general;

    const handleOpen = () => {
        onOpen?.(note);
    };

    return (
        <div
            className={`bg-neutral-2 rounded-lg shadow-outer-1 border-l-4 ${accent} flex flex-col gap-2 p-4 hover:shadow-outer-2 transition-shadow`}
            data-testid={`red-team-note-card-${note.id}`}>
            <div className='flex items-start justify-between gap-2'>
                <div className='flex items-start gap-3 min-w-0'>
                    <div className={`mt-1 ${iconColor}`}>
                        <FontAwesomeIcon icon={NOTE_TYPE_ICONS[note.type] ?? NOTE_TYPE_ICONS.general} />
                    </div>
                    <div className='min-w-0 flex flex-col gap-1'>
                        <div className='flex items-center gap-2 flex-wrap'>
                            <Typography
                                variant='subtitle1'
                                className='font-semibold cursor-pointer hover:underline truncate'
                                onClick={handleOpen}
                                title={note.title}
                                data-testid={`red-team-note-title-${note.id}`}>
                                {note.title}
                            </Typography>
                            <Chip
                                label={NOTE_TYPE_LABELS[note.type] ?? note.type}
                                color={NOTE_TYPE_COLORS[note.type] ?? 'default'}
                                size='small'
                                variant='outlined'
                            />
                        </div>
                        {note.tags.length > 0 && (
                            <div className='flex items-center gap-1 flex-wrap'>
                                {note.tags.map((tag) => (
                                    <Chip key={tag} label={tag} size='small' className='text-xs' />
                                ))}
                            </div>
                        )}
                    </div>
                </div>
                <div className='flex items-center gap-1 shrink-0'>
                    {note.url && (
                        <Tooltip title={note.url}>
                            <IconButton
                                size='small'
                                component='a'
                                href={note.url}
                                target='_blank'
                                rel='noreferrer'
                                data-testid={`red-team-note-link-${note.id}`}>
                                <FontAwesomeIcon icon={faLink} />
                            </IconButton>
                        </Tooltip>
                    )}
                    <Tooltip title='Copy content'>
                        <IconButton
                            size='small'
                            onClick={() => copyToClipboard(note.content)}
                            data-testid={`red-team-note-copy-${note.id}`}>
                            <FontAwesomeIcon icon={faCopy} />
                        </IconButton>
                    </Tooltip>
                    {onEdit && (
                        <Tooltip title='Edit note'>
                            <IconButton
                                size='small'
                                onClick={() => onEdit(note)}
                                data-testid={`red-team-note-edit-${note.id}`}>
                                <FontAwesomeIcon icon={faPen} />
                            </IconButton>
                        </Tooltip>
                    )}
                    {onDelete && (
                        <Tooltip title='Delete note'>
                            <IconButton
                                size='small'
                                onClick={() => onDelete(note)}
                                data-testid={`red-team-note-delete-${note.id}`}>
                                <FontAwesomeIcon icon={faTrash} />
                            </IconButton>
                        </Tooltip>
                    )}
                </div>
            </div>

            {note.content &&
                (compact ? (
                    <Typography
                        variant='body2'
                        className='cursor-pointer opacity-90'
                        onClick={handleOpen}
                        onKeyDown={adaptClickHandlerToKeyDown(handleOpen)}
                        role='button'
                        tabIndex={0}
                        data-testid={`red-team-note-content-${note.id}`}>
                        {markdownExcerpt(note.content)}
                    </Typography>
                ) : (
                    <div
                        role='button'
                        tabIndex={0}
                        className='cursor-pointer text-sm opacity-90'
                        onClick={handleOpen}
                        onKeyDown={adaptClickHandlerToKeyDown(handleOpen)}
                        data-testid={`red-team-note-content-${note.id}`}>
                        <MarkdownContent markdown={note.content} />
                    </div>
                ))}

            <div className='flex items-center gap-3 text-xs opacity-70 flex-wrap'>
                {note.object_id && (
                    <span className='flex items-center gap-1' title='Linked graph object'>
                        <FontAwesomeIcon icon={faCodeBranch} /> {note.object_id}
                    </span>
                )}
                {note.edge_kind && (
                    <span className='flex items-center gap-1' title='Linked edge kind'>
                        edge: {note.edge_kind}
                    </span>
                )}
                <Tooltip title={formatAbsoluteTime(note.updated_at)}>
                    <span className='flex items-center gap-1'>
                        <FontAwesomeIcon icon={faClock} /> {formatRelativeTime(note.updated_at)}
                    </span>
                </Tooltip>
            </div>
        </div>
    );
};

export default NoteCard;
