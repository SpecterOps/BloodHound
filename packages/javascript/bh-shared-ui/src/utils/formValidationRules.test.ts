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

import { act, renderHook } from '@testing-library/react';
import { RegisterOptions, useForm } from 'react-hook-form';
import { describe, expect, it } from 'vitest';
import { descriptionRules, nameRules, requiredRule } from './formValidationRules';

type Values = { value: string };

// Exercise the rules through React Hook Form so required, length, pattern, and
// custom validators run in the same order they do in consuming forms.
async function validate(value: string, rules: RegisterOptions<Values, 'value'>) {
    const { result } = renderHook(() => {
        const form = useForm<Values>({ defaultValues: { value } });
        form.register('value', rules);
        return form;
    });
    await act(async () => {
        await result.current.trigger('value');
    });
    return result.current.getFieldState('value').error?.message;
}

describe('requiredRule', () => {
    it.each(['', '   ', '\t', '\n', ' \t\n '])(
        'rejects blank strings when rejectSpaces is enabled: %#',
        async (value) => {
            expect(await validate(value, requiredRule('Key is required', { rejectSpaces: true }))).toBe(
                'Key is required'
            );
        }
    );

    it.each(['value', ' value', 'value ', 'two words', '\tvalue\n'])(
        'accepts nonblank strings without trimming them: %#',
        async (value) => {
            expect(await validate(value, requiredRule('Key is required', { rejectSpaces: true }))).toBeUndefined();
        }
    );

    it('preserves existing behavior when rejectSpaces is omitted or disabled', async () => {
        expect(await validate('   ', requiredRule('Required'))).toBeUndefined();
        expect(await validate('   ', requiredRule('Required', { rejectSpaces: false }))).toBeUndefined();
    });

    it('reports the supplied message for an empty value', async () => {
        expect(await validate('', requiredRule('Service is required'))).toBe('Service is required');
    });

    it('does not apply string validation to other value types', async () => {
        const { result } = renderHook(() => {
            const form = useForm<{ count: number }>({ defaultValues: { count: 0 } });
            form.register('count', requiredRule('Count is required', { rejectSpaces: true }));
            return form;
        });
        await act(async () => {
            await result.current.trigger('count');
        });
        expect(result.current.getFieldState('count').error).toBeUndefined();
    });
});

describe('nameRules', () => {
    it.each([
        ['', 'Name is required'],
        ['A', 'Name must be 2 characters or more'],
        ['A'.repeat(320), 'Name must be less than 320 characters'],
        ['Plan!', 'Name must be alphanumeric.'],
        [' Plan', 'Name does not allow leading or trailing spaces'],
        ['Plan ', 'Name does not allow leading or trailing spaces'],
        ['   ', 'Name does not allow leading or trailing spaces'],
    ])('rejects invalid name %#', async (value, message) => {
        expect(await validate(value, nameRules())).toBe(message);
    });

    it.each(['AB', 'A'.repeat(319), 'Production AWS 01', 'Production_AWS-01'])(
        'accepts valid name %#',
        async (value) => {
            expect(await validate(value, nameRules())).toBeUndefined();
        }
    );

    it.each([
        ['AB', 'Name must be 3 characters or more'],
        ['ABC', undefined],
        ['ABCDE', undefined],
        ['ABCDEF', 'Name must be less than 6 characters'],
    ])('honors custom length boundaries for %s', async (value, message) => {
        expect(await validate(value, nameRules({ minLength: 3, maxLength: 5 }))).toBe(message);
    });

    it.each(['Plan_Name', 'Plan-Name'])('can disallow underscores and hyphens: %s', async (value) => {
        expect(await validate(value, nameRules({ allowUnderscoreAndHyphen: false }))).toBe(
            'Name must be alphanumeric.'
        );
    });

    it('still accepts spaces and digits when underscores and hyphens are disabled', async () => {
        expect(await validate('Plan 01', nameRules({ allowUnderscoreAndHyphen: false }))).toBeUndefined();
    });

    it('honors a zero minimum instead of using the default', async () => {
        expect(await validate('A', nameRules({ minLength: 0 }))).toBeUndefined();
    });
});

describe('descriptionRules', () => {
    it.each(['', 'A description, with punctuation!', 'A'.repeat(500)])(
        'accepts valid description %#',
        async (value) => {
            expect(await validate(value, descriptionRules())).toBeUndefined();
        }
    );

    it('enforces the default maximum', async () => {
        expect(await validate('A'.repeat(501), descriptionRules())).toBe(
            'Description must be less than 501 characters'
        );
    });

    it.each([' Description', 'Description ', '   '])('rejects surrounding whitespace %#', async (value) => {
        expect(await validate(value, descriptionRules())).toBe('Description does not allow leading or trailing spaces');
    });

    it('honors a custom maximum at its boundary', async () => {
        expect(await validate('12345', descriptionRules({ maxLength: 5 }))).toBeUndefined();
        expect(await validate('123456', descriptionRules({ maxLength: 5 }))).toBe(
            'Description must be less than 6 characters'
        );
    });

    it('honors a zero maximum instead of using the default', async () => {
        expect(await validate('A', descriptionRules({ maxLength: 0 }))).toBe(
            'Description must be less than 1 characters'
        );
    });
});
