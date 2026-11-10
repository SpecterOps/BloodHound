// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

import { faCopy } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton, Menu, MenuContent, MenuItem, MenuTrigger } from 'doodle-ui';
import { DateTime } from 'luxon';
import { useAnnounce, useNotifications } from '../../providers';
import { copyToClipboard, DATE_FIELDS, EntityField } from '../../utils';

export const serializeRawPropertyValue = (value: EntityField['value']): string => {
    if (typeof value === 'string') return value;
    if (typeof value === 'number') return Object.is(value, -0) ? '-0' : String(value);
    if (typeof value === 'boolean' || value === null) return String(value);

    return JSON.stringify(value);
};

export const serializeISO8601PropertyValue = (property: string, value: EntityField['value']): string | null => {
    if (typeof value === 'string') {
        const date = DateTime.fromISO(value, { setZone: true });
        return date.isValid ? date.toUTC().toISO() : null;
    }

    if (typeof value !== 'number' || !DATE_FIELDS.includes(property) || value <= 0) return null;

    return DateTime.fromSeconds(value).toUTC().toISO();
};

export const EntityPropertyCopyMenu = ({ property, value }: { property?: string; value: EntityField['value'] }) => {
    const announce = useAnnounce();
    const { addNotification } = useNotifications();

    if (!property) return null;

    const iso8601Value = property ? serializeISO8601PropertyValue(property, value) : null;
    const copy = async (kind: 'field name' | 'raw value' | 'ISO 8601 value', text: string) => {
        try {
            await copyToClipboard(text);
            const message = `${kind === 'field name' ? 'Field name' : kind === 'raw value' ? 'Raw value' : 'ISO 8601 value'} copied to clipboard`;
            addNotification(message, `copy-property-${kind.replace(' ', '-')}`);
            announce(message);
        } catch {
            const message = `Unable to copy ${kind}`;
            addNotification(message, `copy-property-${kind.replace(' ', '-')}-error`);
            announce(message, 'assertive');
        }
    };

    return (
        <Menu>
            <MenuTrigger asChild>
                <IconButton
                    aria-label={`Copy ${property}`}
                    hideTooltip
                    className='ml-1 size-6 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100 focus-visible:opacity-100'>
                    <FontAwesomeIcon icon={faCopy} />
                </IconButton>
            </MenuTrigger>
            <MenuContent aria-label={`Copy options for ${property}`} align='start'>
                <MenuItem onSelect={() => void copy('field name', property)}>Copy field name</MenuItem>
                <MenuItem onSelect={() => void copy('raw value', serializeRawPropertyValue(value))}>
                    Copy raw value
                </MenuItem>
                {iso8601Value && (
                    <MenuItem onSelect={() => void copy('ISO 8601 value', iso8601Value)}>Copy ISO 8601 value</MenuItem>
                )}
            </MenuContent>
        </Menu>
    );
};
