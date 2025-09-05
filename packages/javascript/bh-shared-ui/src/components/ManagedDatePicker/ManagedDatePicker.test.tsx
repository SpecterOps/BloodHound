// Copyright 2025 Specter Ops, Inc.
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
import { render, screen } from '../../test-utils';
import { ManagedDatePicker, VALIDATIONS } from './ManagedDatePicker';

const JAN_1 = new Date('2025-01-01T00:00:00Z');
const JAN_2 = new Date('2025-01-02T00:00:00Z');
const JAN_31 = new Date('2025-01-31T00:00:00Z');

const originalTZ = process.env.TZ;

const DEFAULT_VALIDATIONS = [VALIDATIONS.isBeforeDate(JAN_2, 'Date is not before Jan 2nd')];

type RenderDatePickeOptions = {
    hint?: string;
    value?: Date;
    validations?: typeof DEFAULT_VALIDATIONS;
};

const renderDatePicker = (props: RenderDatePickeOptions = {}) => {
    const { hint, value, validations } = props;
    const user = userEvent.setup();

    const onDateChangeMock = vi.fn();
    const onValidationMock = vi.fn();

    render(
        <ManagedDatePicker
            fromDate={JAN_1}
            hint={hint}
            onDateChange={onDateChangeMock}
            onValidation={onValidationMock}
            toDate={JAN_31}
            validations={validations}
            value={value}
        />
    );

    return {
        clickAway: () => user.click(document.body),
        clickInput: () => user.click(screen.getByRole('textbox')),
        onDateChangeMock,
        onValidationMock,
        openCalendar: () => user.click(screen.getByRole('button', { name: /choose date/i })),
        selectDate: (date: string) => user.click(screen.getAllByRole('gridcell', { name: date })[0]),
        typeInput: (text: string) => user.type(screen.getByRole('textbox'), text),
    };
};

beforeAll(() => {
    process.env.TZ = 'UTC';
});

afterAll(() => {
    process.env.TZ = originalTZ;
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe('VALIDATIONS', () => {
    it('returns true when isAfterDate is called without after or date', () => {
        expect(VALIDATIONS.isAfterDate(undefined, '').rule(JAN_2)).toBeTruthy();
    });

    it('returns true when date is after or equal to the given after date', () => {
        expect(VALIDATIONS.isAfterDate(JAN_1, '').rule(JAN_2)).toBeTruthy();
    });

    it('returns false when date is before the given after date', () => {
        expect(VALIDATIONS.isAfterDate(JAN_2, '').rule(JAN_1)).toBeFalsy();
    });

    it('returns true when isBeforeDate is called without before or date', () => {
        expect(VALIDATIONS.isBeforeDate(undefined, '').rule(JAN_2)).toBeTruthy();
    });

    it('returns true when date is before or equal to the given before date', () => {
        expect(VALIDATIONS.isBeforeDate(JAN_2, '').rule(JAN_1)).toBeTruthy();
    });

    it('returns false when date is after the given before date', () => {
        expect(VALIDATIONS.isBeforeDate(JAN_1, '').rule(JAN_2)).toBeFalsy();
    });
});

describe('ManagedDatePicker - date validation', () => {
    it('returns true when no validations are provided', async () => {
        const { clickAway, onValidationMock, openCalendar, selectDate } = renderDatePicker();

        await openCalendar();
        await selectDate('1');

        expect(onValidationMock).toHaveBeenCalledWith(true);

        clickAway();

        await openCalendar();
        await selectDate('3');

        expect(onValidationMock).toHaveBeenCalledWith(true);
    });

    it('returns true when validations pass', async () => {
        const { onValidationMock, openCalendar, selectDate } = renderDatePicker({ validations: DEFAULT_VALIDATIONS });

        await openCalendar();
        await selectDate('1');

        expect(onValidationMock).toHaveBeenCalledWith(true);
    });

    it('returns false when validations fails', async () => {
        const { onValidationMock, openCalendar, selectDate } = renderDatePicker({ validations: DEFAULT_VALIDATIONS });

        await openCalendar();
        await selectDate('3');

        expect(onValidationMock).toHaveBeenCalledWith(false);
    });

    it('sets input text on date click', async () => {
        const { openCalendar, selectDate } = renderDatePicker();

        await openCalendar();
        await selectDate('3');

        expect(screen.getByRole('textbox')).toHaveValue('2025-01-03');
    });

    it('displays a validation error', async () => {
        const { openCalendar, selectDate } = renderDatePicker({ validations: DEFAULT_VALIDATIONS });

        await openCalendar();
        await selectDate('3');

        expect(screen.getByText('Date is not before Jan 2nd')).toBeInTheDocument();
    });

    it('clear validation error on input', async () => {
        const { clickAway, typeInput } = renderDatePicker({ validations: DEFAULT_VALIDATIONS });

        await typeInput('2025-01-03');
        await clickAway();

        expect(screen.getByText('Date is not before Jan 2nd')).toBeInTheDocument();

        await typeInput('{backspace}'.repeat(1));

        screen.logTestingPlaygroundURL();

        expect(screen.queryByText('Date is not before Jan 2nd')).not.toBeInTheDocument();
    });
});

describe('ManagedDatePicker - placeholders', () => {
    it('shows date placeholder if no hint is given', async () => {
        renderDatePicker();

        expect(screen.getByPlaceholderText(/yyyy-mm-dd/i)).toBeInTheDocument();
    });

    it('shows hint by default', async () => {
        renderDatePicker({ hint: 'a hint' });

        expect(screen.getByPlaceholderText(/a hint/i)).toBeInTheDocument();
    });

    it('if a hint is given, it show date placeholder on focus, restores hint on blur', async () => {
        const { clickAway, clickInput } = renderDatePicker({ hint: 'a hint' });

        await clickInput();
        expect(screen.getByPlaceholderText(/yyyy-mm-dd/i)).toBeInTheDocument();

        await clickAway();
        expect(screen.getByPlaceholderText(/a hint/i)).toBeInTheDocument();
    });
});

describe('ManagedDatePicker - value', () => {
    it('does not set input text if no value is given', async () => {
        renderDatePicker();

        expect(screen.getByRole('textbox')).toHaveValue('');
    });

    it('sets input text to value', async () => {
        renderDatePicker({ value: JAN_1 });

        expect(screen.getByRole('textbox')).toHaveValue('2025-01-01');
    });

    it('updates value after blur when a valid date typed', async () => {
        const { clickAway, onDateChangeMock, typeInput } = renderDatePicker();

        await typeInput('2025-01-01');
        await clickAway();

        expect(onDateChangeMock).toHaveBeenCalledWith(new Date('2025-01-01T00:00:00.000Z'));
    });

    it('does not updates value after blur when bad date typed', async () => {
        const { clickAway, onDateChangeMock, typeInput } = renderDatePicker();

        await typeInput('2025');
        await clickAway();

        expect(onDateChangeMock).not.toHaveBeenCalled();
    });
});

describe('ManagedDatePicker - input', () => {
    it('updates calendar when a valid date typed', async () => {
        const { openCalendar, typeInput } = renderDatePicker();

        await typeInput('2025-01-01');
        await openCalendar();

        const calendarDay = screen.getByRole('gridcell', { selected: true });

        expect(calendarDay).toBeInTheDocument();
        expect(calendarDay).toHaveTextContent('1');
    });

    it('does not updates calendar when bad date typed', async () => {
        const { openCalendar, typeInput } = renderDatePicker();

        await typeInput('2025');
        await openCalendar();

        const calendarDay = screen.queryByRole('gridcell', { selected: true });

        expect(calendarDay).not.toBeInTheDocument();
    });

    it('clears value if input is empty', async () => {
        const { clickAway, onDateChangeMock, typeInput } = renderDatePicker();

        await typeInput('2025-01-01');
        await clickAway();

        expect(onDateChangeMock).toHaveBeenCalledWith(new Date('2025-01-01T00:00:00.000Z'));

        await typeInput('{backspace}'.repeat(10));

        await clickAway();

        expect(onDateChangeMock).toHaveBeenCalledWith();
    });
});
