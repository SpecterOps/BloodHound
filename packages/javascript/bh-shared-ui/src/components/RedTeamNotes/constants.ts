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

import { faCode, faCrosshairs, faNoteSticky, faWrench, IconDefinition } from '@fortawesome/free-solid-svg-icons';
import { RedTeamNoteType } from 'js-client-library';

export const NOTE_TYPE_LABELS: Record<RedTeamNoteType, string> = {
    general: 'General',
    technique: 'Technique',
    tool: 'Tool',
    source: 'Source',
};

export const NOTE_TYPE_OPTIONS: { value: RedTeamNoteType; label: string }[] = [
    { value: 'general', label: 'General' },
    { value: 'technique', label: 'Technique' },
    { value: 'tool', label: 'Tool' },
    { value: 'source', label: 'Source' },
];

export const NOTE_TYPE_COLORS: Record<
    RedTeamNoteType,
    'default' | 'primary' | 'secondary' | 'error' | 'info' | 'success' | 'warning'
> = {
    general: 'default',
    technique: 'error',
    tool: 'info',
    source: 'success',
};

export const NOTE_TYPE_ICONS: Record<RedTeamNoteType, IconDefinition> = {
    general: faNoteSticky,
    technique: faCrosshairs,
    tool: faWrench,
    source: faCode,
};

export const NOTE_TYPE_DESCRIPTIONS: Record<RedTeamNoteType, string> = {
    general: 'Free-form red team knowledge',
    technique: 'Attack technique walkthroughs and tradecraft',
    tool: 'Offensive tooling usage and references',
    source: 'Source code, repos and research material',
};

export const NOTE_SORT_OPTIONS: { value: string; label: string }[] = [
    { value: '-updated_at', label: 'Recently updated' },
    { value: 'updated_at', label: 'Oldest updated' },
    { value: 'title', label: 'Title A-Z' },
    { value: '-title', label: 'Title Z-A' },
];
