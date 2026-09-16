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
    it.each([' value', 'value ', '   ', '\tvalue', 'value\n'])(
        'rejects surrounding whitespace when enabled: %#',
        async (value) => {
            expect(
                await validate(
                    value,
                    requiredRule('Key is required', {
                        shouldRejectSurroundingWhitespace: true,
                        label: 'Key',
                    })
                )
            ).toBe('Key does not allow leading or trailing spaces');
        }
    );

    it('preserves the required message for an empty value with whitespace validation enabled', async () => {
        expect(
            await validate(
                '',
                requiredRule('Key is required', {
                    shouldRejectSurroundingWhitespace: true,
                })
            )
        ).toBe('Key is required');
    });

    it('uses a default label when none is supplied', async () => {
        expect(
            await validate(
                ' value ',
                requiredRule('Required', {
                    shouldRejectSurroundingWhitespace: true,
                })
            )
        ).toBe('Field does not allow leading or trailing spaces');
    });

    it('allows interior spaces with whitespace validation enabled', async () => {
        expect(
            await validate(
                'two words',
                requiredRule('Required', {
                    shouldRejectSurroundingWhitespace: true,
                })
            )
        ).toBeUndefined();
    });

    it('preserves existing behavior when whitespace validation is omitted or disabled', async () => {
        expect(await validate(' value ', requiredRule('Required'))).toBeUndefined();
        expect(
            await validate(
                ' value ',
                requiredRule('Required', {
                    shouldRejectSurroundingWhitespace: false,
                })
            )
        ).toBeUndefined();
    });

    it('reports the supplied message for an empty value', async () => {
        expect(await validate('', requiredRule('Service is required'))).toBe('Service is required');
    });

    it('accepts a populated value', async () => {
        expect(await validate('aws', requiredRule('Service is required'))).toBeUndefined();
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
