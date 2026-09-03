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

import JSZip from 'jszip';
import { RedTeamNote, RedTeamNoteAttachment, RedTeamNotePayload } from 'js-client-library';

const MEDIA_URL_PATTERN = /\/api\/v2\/red-team-notes\/media\/([0-9a-fA-F-]+)/g;
const RELATIVE_ATTACHMENT_PATTERN = /attachments\/([^\s)]+)/g;

const extensionForContentType = (contentType: string): string => {
    switch (contentType) {
        case 'image/jpeg':
            return 'jpg';
        case 'image/gif':
            return 'gif';
        case 'image/webp':
            return 'webp';
        case 'image/svg+xml':
            return 'svg';
        case 'image/png':
        default:
            return 'png';
    }
};

const contentTypeForExtension = (name: string): string => {
    if (name.endsWith('.jpg') || name.endsWith('.jpeg')) return 'image/jpeg';
    if (name.endsWith('.gif')) return 'image/gif';
    if (name.endsWith('.webp')) return 'image/webp';
    if (name.endsWith('.svg')) return 'image/svg+xml';
    return 'image/png';
};

export const slugify = (title: string): string =>
    title
        .toLowerCase()
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .replace(/đ/g, 'd')
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '')
        .slice(0, 60) || 'note';

const buildFrontmatter = (note: RedTeamNote): string => {
    const lines = [
        '---',
        `title: ${JSON.stringify(note.title)}`,
        `type: ${JSON.stringify(note.type)}`,
        `tags: ${JSON.stringify(note.tags)}`,
        `url: ${JSON.stringify(note.url)}`,
        `object_id: ${JSON.stringify(note.object_id)}`,
        `edge_kind: ${JSON.stringify(note.edge_kind)}`,
        `created_at: ${JSON.stringify(note.created_at)}`,
        `updated_at: ${JSON.stringify(note.updated_at)}`,
        '---',
        '',
    ];
    return lines.join('\n');
};

export const parseFrontmatter = (raw: string): { fields: Record<string, any>; content: string } => {
    if (!raw.startsWith('---')) {
        return { fields: {}, content: raw };
    }

    const end = raw.indexOf('\n---', 3);
    if (end === -1) {
        return { fields: {}, content: raw };
    }

    const header = raw.slice(4, end);
    const content = raw.slice(end + 4).replace(/^\n/, '');
    const fields: Record<string, any> = {};

    for (const line of header.split('\n')) {
        const separator = line.indexOf(':');
        if (separator === -1) continue;
        const key = line.slice(0, separator).trim();
        const value = line.slice(separator + 1).trim();
        try {
            fields[key] = JSON.parse(value);
        } catch {
            fields[key] = value;
        }
    }

    return { fields, content };
};

export interface ExportResult {
    noteCount: number;
    attachmentCount: number;
}

// exportNotesZip bundles every note as a markdown file (YAML frontmatter +
// content) plus the referenced attachments, rewriting media URLs to relative
// paths so the archive is self-contained and readable outside BloodHound.
export const exportNotesZip = async (
    notes: RedTeamNote[],
    fetchMedia: (token: string) => Promise<Response>
): Promise<{ blob: Blob; stats: ExportResult }> => {
    const zip = new JSZip();
    let attachmentCount = 0;

    for (const note of notes) {
        let content = note.content;
        const tokens = Array.from(new Set(Array.from(content.matchAll(MEDIA_URL_PATTERN)).map((match) => match[1])));

        for (const token of tokens) {
            try {
                const response = await fetchMedia(token);
                if (!response.ok) continue;
                const mediaBlob = await response.blob();
                const fileName = `${token}.${extensionForContentType(mediaBlob.type)}`;
                zip.file(`attachments/${fileName}`, mediaBlob);
                content = content.split(`/api/v2/red-team-notes/media/${token}`).join(`attachments/${fileName}`);
                attachmentCount += 1;
            } catch {
                // keep the original URL when the attachment cannot be fetched
            }
        }

        zip.file(`notes/${slugify(note.title)}-${note.id}.md`, buildFrontmatter(note) + content);
    }

    zip.file(
        'manifest.json',
        JSON.stringify(
            {
                exported_at: new Date().toISOString(),
                source: 'BloodHound Red Team Knowledge Base',
                note_count: notes.length,
                attachment_count: attachmentCount,
                notes: notes.map((note) => ({ id: note.id, title: note.title, type: note.type, tags: note.tags })),
            },
            null,
            2
        )
    );

    const blob = await zip.generateAsync({ type: 'blob' });
    return { blob, stats: { noteCount: notes.length, attachmentCount } };
};

export interface ImportResult {
    createdNotes: number;
    createdAttachments: number;
}

// importNotesZip reads a previously exported archive (or any zip of markdown
// files under notes/), re-uploads bundled attachments and recreates the notes.
export const importNotesZip = async (
    file: File,
    uploadAttachment: (file: File) => Promise<RedTeamNoteAttachment>,
    createNote: (payload: RedTeamNotePayload) => Promise<unknown>
): Promise<ImportResult> => {
    const zip = await JSZip.loadAsync(file);
    const result: ImportResult = { createdNotes: 0, createdAttachments: 0 };

    const notePaths = Object.keys(zip.files)
        .filter((path) => path.startsWith('notes/') && path.endsWith('.md') && !zip.files[path].dir)
        .sort();

    const loosePaths = Object.keys(zip.files).filter(
        (path) => !path.startsWith('notes/') && !path.startsWith('attachments/') && path.endsWith('.md') && !zip.files[path].dir
    );

    for (const path of [...notePaths, ...loosePaths]) {
        const raw = await zip.files[path].async('string');
        const { fields, content: parsedContent } = parseFrontmatter(raw);

        let content = parsedContent;
        const attachmentRefs = Array.from(
            new Set(Array.from(parsedContent.matchAll(RELATIVE_ATTACHMENT_PATTERN)).map((match) => match[1]))
        );

        for (const name of attachmentRefs) {
            const attachmentEntry = zip.files[`attachments/${name}`];
            if (!attachmentEntry) continue;
            const attachmentBlob = await attachmentEntry.async('blob');
            const attachmentFile = new File([attachmentBlob], name, { type: contentTypeForExtension(name) });
            const uploaded = await uploadAttachment(attachmentFile);
            content = content.split(`attachments/${name}`).join(uploaded.url);
            result.createdAttachments += 1;
        }

        const payload: RedTeamNotePayload = {
            title: typeof fields.title === 'string' && fields.title ? fields.title : path.replace(/\.md$/, ''),
            content,
            type: fields.type ?? 'general',
            tags: Array.isArray(fields.tags) ? fields.tags : [],
            url: fields.url ?? '',
            object_id: fields.object_id ?? '',
            edge_kind: fields.edge_kind ?? '',
        };

        await createNote(payload);
        result.createdNotes += 1;
    }

    return result;
};
