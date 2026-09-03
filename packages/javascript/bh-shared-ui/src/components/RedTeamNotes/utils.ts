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

export const formatRelativeTime = (isoDate: string): string => {
    const date = new Date(isoDate);

    if (Number.isNaN(date.getTime())) {
        return isoDate;
    }

    const elapsedSeconds = Math.floor((Date.now() - date.getTime()) / 1000);

    if (elapsedSeconds < 60) {
        return 'just now';
    }

    const elapsedMinutes = Math.floor(elapsedSeconds / 60);
    if (elapsedMinutes < 60) {
        return `${elapsedMinutes}m ago`;
    }

    const elapsedHours = Math.floor(elapsedMinutes / 60);
    if (elapsedHours < 24) {
        return `${elapsedHours}h ago`;
    }

    const elapsedDays = Math.floor(elapsedHours / 24);
    if (elapsedDays < 30) {
        return `${elapsedDays}d ago`;
    }

    return date.toLocaleDateString();
};

export const formatAbsoluteTime = (isoDate: string): string => {
    const date = new Date(isoDate);

    if (Number.isNaN(date.getTime())) {
        return isoDate;
    }

    return date.toLocaleString();
};

export const downloadJsonFile = (filename: string, payload: unknown): void => {
    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
    const objectUrl = URL.createObjectURL(blob);
    const anchor = document.createElement('a');

    anchor.href = objectUrl;
    anchor.download = filename;
    anchor.click();

    URL.revokeObjectURL(objectUrl);
};

// markdownExcerpt flattens markdown into plain text and truncates it, so list
// cards can show a stable, bounded preview regardless of content length.
export const markdownExcerpt = (markdown: string, maxLength = 280): string => {
    const plainText = markdown
        .replace(/```[\s\S]*?```/g, (block) => ' ' + block.replace(/```[a-zA-Z]*\n?/g, '').replace(/```/g, ' ') + ' ')
        .replace(/!\[([^\]]*)\]\([^)]*\)/g, '$1')
        .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
        .replace(/^[#>*+-]+\s*/gm, '')
        .replace(/[*_`~]/g, '')
        .replace(/\s+/g, ' ')
        .trim();

    if (plainText.length <= maxLength) {
        return plainText;
    }

    return plainText.slice(0, maxLength).trimEnd() + '…';
};
