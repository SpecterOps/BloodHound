// Copyright 2023 Specter Ops, Inc.
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

import userEvent from '@testing-library/user-event';
import {
    serializeISO8601PropertyValue,
    serializeRawPropertyValue,
} from '../../components/EntityInfo/EntityPropertyCopyMenu';
import { render, screen } from '../../test-utils';
import { EntityField } from '../../utils';
import { Field, ObjectInfoFields, exclusionList } from './fragments';

describe('Field', () => {
    it.each([
        { description: 'single-item', values: ['HOST/one'] },
        { description: 'multiple-item', values: ['HOST/one', 'HOST/two'] },
    ])('renders $description lists with one label and a title for each value', ({ values }) => {
        render(<Field label='Service Principal Names:' value={values} />);

        expect(screen.getAllByText(/^Service Principal Names:?$/)).toHaveLength(1);
        values.forEach((value) => {
            const element = screen.getByText(value);
            expect(element).toHaveAttribute('title', value);
        });
    });

    it('preserves rich labels', () => {
        render(<Field label={<span>Rich label</span>} value='value' />);
        expect(screen.getByText('Rich label')).toBeInTheDocument();
        expect(screen.getByText('value')).toBeInTheDocument();
    });

    it('should render a Field when the provided value is false', () => {
        render(<Field label='Test Field (Boolean)' value={false} keyprop='test-field-boolean' />);
        expect(screen.getByText('FALSE')).toBeInTheDocument();
    });

    it('should render a Field when the provided value is 0', () => {
        render(<Field label='Test Field (Number)' value={0} keyprop='test-field-number' />);
        expect(screen.getByText('0')).toBeInTheDocument();
    });

    it('should render a Field when the provided value is -0', () => {
        render(<Field label='Test Field (Number)' value={-0} keyprop='test-field-number' />);
        expect(screen.getByText('0')).toBeInTheDocument();
        expect(screen.queryByText('-0')).not.toBeInTheDocument();
    });

    it('should render an accessible placeholder when the provided value is an empty string', () => {
        render(<Field label='Test Field (String)' value={''} keyprop='test-field-string' />);
        expect(screen.getByText('Test Field (String)')).toBeInTheDocument();
        expect(screen.getByText('—')).toHaveAttribute('aria-hidden', 'true');
        expect(screen.getByText('Empty string')).toBeInTheDocument();
    });

    it('should render an accessible placeholder when the provided value is an empty array', () => {
        render(<Field label='Test Field (Array)' value={[]} keyprop='test-field-array' />);
        expect(screen.getByText('Test Field (Array)')).toBeInTheDocument();
        expect(screen.getByText('NONE')).toHaveAttribute('aria-hidden', 'true');
        expect(screen.getByText('Empty array, zero values')).toBeInTheDocument();
    });

    it('should render accessible placeholders for empty values in a list', () => {
        render(<Field label='Test Field (Array)' value={['value', '', null]} keyprop='test-field-array' />);
        expect(screen.getByText('value')).not.toHaveAttribute('aria-hidden');
        screen.getAllByText('—').forEach((placeholder) => {
            expect(placeholder).toHaveAttribute('aria-hidden', 'true');
        });
        expect(screen.getByText('Empty string')).toBeInTheDocument();
        expect(screen.getByText('Null value')).toBeInTheDocument();
    });

    it('should render an accessible placeholder when the provided value is null', () => {
        render(<Field label='Test Field (Null)' value={null} keyprop='test-field-null' />);
        expect(screen.getByText('Test Field (Null)')).toBeInTheDocument();
        expect(screen.getByText('—')).toHaveAttribute('aria-hidden', 'true');
        expect(screen.getByText('Null value')).toBeInTheDocument();
    });

    it('should render a real value matching an empty state placeholder as ordinary text', () => {
        render(<Field label='Test Field (String)' value='NONE' keyprop='test-field-string' />);
        expect(screen.getByText('NONE')).not.toHaveAttribute('aria-hidden');
        expect(screen.queryByText('Empty array, zero values')).not.toBeInTheDocument();
    });

    it('copies the raw property key and value rather than their display formatting', async () => {
        const user = userEvent.setup();
        render(<Field label='Last logon:' value={0} keyprop='lastlogon' />);

        await user.click(screen.getByRole('button', { name: 'Copy lastlogon' }));
        await user.click(screen.getByRole('menuitem', { name: 'Copy field name' }));
        expect(await navigator.clipboard.readText()).toBe('lastlogon');

        await user.click(screen.getByRole('button', { name: 'Copy lastlogon' }));
        await user.click(screen.getByRole('menuitem', { name: 'Copy raw value' }));
        expect(await navigator.clipboard.readText()).toBe('0');
    });

    it('supports object properties by displaying and copying their JSON value', async () => {
        const user = userEvent.setup();
        render(<Field label='Custom property:' value={{ nested: ['value'] }} keyprop='customProperty' />);

        expect(screen.getByText('{"nested":["value"]}')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: 'Copy customProperty' }));
        await user.click(screen.getByRole('menuitem', { name: 'Copy raw value' }));
        expect(await navigator.clipboard.readText()).toBe('{"nested":["value"]}');
    });

    it('announces clipboard failures', async () => {
        const user = userEvent.setup();
        const writeText = vi.spyOn(navigator.clipboard, 'writeText').mockRejectedValueOnce(new Error('denied'));
        render(<Field label='Test field:' value='value' keyprop='testfield' />);

        await user.click(screen.getByRole('button', { name: 'Copy testfield' }));
        await user.click(screen.getByRole('menuitem', { name: 'Copy raw value' }));

        expect(await screen.findByText('Unable to copy raw value')).toBeInTheDocument();
        writeText.mockRestore();
    });

    it('opens the copy menu with the keyboard', async () => {
        const user = userEvent.setup();
        render(<Field label='Test field:' value='value' keyprop='testfield' />);

        const trigger = screen.getByRole('button', { name: 'Copy testfield' });
        trigger.focus();
        await user.keyboard('{Enter}');

        expect(screen.getByRole('menuitem', { name: 'Copy field name' })).toBeInTheDocument();
        expect(screen.getByRole('menuitem', { name: 'Copy raw value' })).toBeInTheDocument();
    });

    it('offers ISO 8601 copying for date properties', async () => {
        const user = userEvent.setup();
        render(<Field label='Last logon:' value={1694549003} keyprop='lastlogon' />);

        await user.click(screen.getByRole('button', { name: 'Copy lastlogon' }));
        await user.click(screen.getByRole('menuitem', { name: 'Copy ISO 8601 value' }));

        expect(await navigator.clipboard.readText()).toBe('2023-09-12T20:03:23.000Z');
    });
});

describe('serializeRawPropertyValue', () => {
    it.each([
        [false, 'false'],
        [null, 'null'],
        ['', ''],
        [[], '[]'],
        [{ nested: ['value'] }, '{"nested":["value"]}'],
    ])('serializes %j as an underlying clipboard value', (value, expected) => {
        expect(serializeRawPropertyValue(value as EntityField['value'])).toBe(expected);
    });
});

describe('serializeISO8601PropertyValue', () => {
    it('serializes date strings and timestamp date fields as UTC ISO 8601', () => {
        expect(serializeISO8601PropertyValue('customDate', '2023-09-12T20:03:23-07:00')).toBe(
            '2023-09-13T03:03:23.000Z'
        );
        expect(serializeISO8601PropertyValue('lastlogon', 1694549003)).toBe('2023-09-12T20:03:23.000Z');
    });

    it('does not offer ISO 8601 for non-date values or date sentinels', () => {
        expect(serializeISO8601PropertyValue('name', 'not-a-date')).toBeNull();
        expect(serializeISO8601PropertyValue('lastlogon', 0)).toBeNull();
    });
});

describe('ObjectInfoFields', () => {
    it('should filter properties that we do not want to display in the entity panel', () => {
        const properties: EntityField[] = exclusionList.map((key) => {
            return { label: key, value: 'test', keyprop: key };
        });
        const validProperty: EntityField = { label: 'Object ID', value: 'OBJECT_ID' };
        properties.push(validProperty);

        render(<ObjectInfoFields fields={properties} />);

        expect(screen.getByText('Object ID')).toBeInTheDocument();

        exclusionList.forEach((property) => {
            expect(screen.queryByText(property)).not.toBeInTheDocument();
        });

        expect(screen.queryByRole('button', { name: 'Copy undefined' })).not.toBeInTheDocument();
    });
});
