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

export const CARD_ACCENTS = ['blue', 'orange'] as const;
export const CARD_ICONS = ['storybook', 'documentation'] as const;

export type CardAccent = (typeof CARD_ACCENTS)[number];
export type CardIcon = (typeof CARD_ICONS)[number];

export type PortalCardConfig = {
    accent: CardAccent;
    title: string;
    description: string;
    href: string;
    icon: CardIcon;
};

export type DocsConfig = {
    defaultPage: string;
    repositoryUrl: string;
    include: string[];
    exclude: string[];
};

export type PortalConfig = {
    cards: PortalCardConfig[];
    docs: DocsConfig;
};

type UnknownRecord = Record<string, unknown>;

/**
 * Validates a single card entry, appending any problems found to `problems`.
 */
const checkCard = (card: unknown, location: string, problems: string[]) => {
    if (!isRecord(card)) {
        problems.push(`${location} must be an object`);
        return;
    }
    checkKeys(card, ['accent', 'title', 'description', 'href', 'icon'], location, problems);
    ['title', 'description', 'href'].forEach((key) => {
        if (!isNonEmptyString(card[key])) problems.push(`${location}.${key} must be a non-empty string`);
    });
    if (!CARD_ACCENTS.includes(card.accent as CardAccent)) {
        problems.push(`${location}.accent must be one of: ${CARD_ACCENTS.join(', ')}`);
    }
    if (!CARD_ICONS.includes(card.icon as CardIcon)) {
        problems.push(`${location}.icon must be one of: ${CARD_ICONS.join(', ')}`);
    }
};

/**
 * Validates the `docs` section, appending any problems found to `problems`.
 */
const checkDocs = (docs: unknown, problems: string[]) => {
    if (!isRecord(docs)) {
        problems.push('docs must be an object');
        return;
    }
    checkKeys(docs, ['defaultPage', 'repositoryUrl', 'include', 'exclude'], 'docs', problems);
    if (!isNonEmptyString(docs.defaultPage)) problems.push('docs.defaultPage must be a non-empty string');
    if (!isNonEmptyString(docs.repositoryUrl) || !/^https?:\/\/.+\/$/.test(docs.repositoryUrl)) {
        problems.push('docs.repositoryUrl must be an http(s) URL ending in "/"');
    }
    if (!isStringList(docs.include) || docs.include.length === 0) {
        problems.push('docs.include must be a non-empty list of glob patterns');
    }
    if (!isStringList(docs.exclude)) problems.push('docs.exclude must be a list of glob patterns');
};

/**
 * Reports every key in `record` that is not in `allowedKeys` as a problem.
 */
const checkKeys = (record: UnknownRecord, allowedKeys: string[], location: string, problems: string[]) => {
    Object.keys(record)
        .filter((key) => !allowedKeys.includes(key))
        .forEach((key) => problems.push(`${location}.${key} is not a known setting`));
};

/**
 * Returns true when `value` is a string containing non-whitespace characters.
 */
const isNonEmptyString = (value: unknown): value is string => typeof value === 'string' && value.trim() !== '';

/**
 * Returns true when `value` is a plain object (not null or an array).
 */
const isRecord = (value: unknown): value is UnknownRecord =>
    typeof value === 'object' && value !== null && !Array.isArray(value);

/**
 * Returns true when `value` is an array of non-empty strings.
 */
const isStringList = (value: unknown): value is string[] => Array.isArray(value) && value.every(isNonEmptyString);

/**
 * Validates the parsed contents of conf.json, reporting every problem at once.
 */
export const parsePortalConfig = (raw: unknown): PortalConfig => {
    const problems: string[] = [];

    if (!isRecord(raw)) {
        problems.push('the config must be a JSON object');
    } else {
        checkKeys(raw, ['cards', 'docs'], 'config', problems);
        if (!Array.isArray(raw.cards)) {
            problems.push('cards must be a list');
        } else {
            raw.cards.forEach((card, index) => checkCard(card, `cards[${index}]`, problems));
        }
        checkDocs(raw.docs, problems);
    }

    if (problems.length > 0) {
        throw new Error(`Invalid Dev Portal config:\n${problems.map((problem) => `  - ${problem}`).join('\n')}`);
    }

    return raw as PortalConfig;
};
