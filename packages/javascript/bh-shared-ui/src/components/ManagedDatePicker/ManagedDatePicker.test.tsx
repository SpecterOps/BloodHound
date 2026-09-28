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
import { ManagedDatePicker } from './ManagedDatePicker';

const JAN_1 = new Date('2025-01-01T00:00:00Z');
const JAN_31 = new Date('2025-01-31T00:00:00Z');

const originalTZ = process.env.TZ;

type RenderDatePickerOptions = {
    hint?: string;
    value?: string;
};

const renderDatePicker = (props: RenderDatePickerOptions = {}) => {
    const { hint, value } = props;
    const user = userEvent.setup();

    const onDateChangeMock = vi.fn();

    render(
        <ManagedDatePicker fromDate={JAN_1} hint={hint} onDateChange={onDateChangeMock} toDate={JAN_31} value={value} />
    );

    return {
        clickAway: () => user.click(document.body),
        clickInput: () => user.click(screen.getByRole('textbox')),
        onDateChangeMock,
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

describe('ManagedDatePicker - date validation', () => {
    it('sets input text on date click', async () => {
        const { openCalendar, selectDate } = renderDatePicker();

        await openCalendar();
        await selectDate('3');

        expect(screen.getByRole('textbox')).toHaveValue('2025-01-03');
    });

    it('displays a validation error', async () => {
        const { typeInput } = renderDatePicker();

        await typeInput('not a date');

        expect(await screen.findByText('Input is not a valid date.')).toBeInTheDocument();
    });

    it('clear validation error on input', async () => {
        const { clickAway, typeInput } = renderDatePicker();

        await typeInput('not a date');

        expect(await screen.findByText('Input is not a valid date.')).toBeInTheDocument();

        await typeInput('{backspace}'.repeat(10));
        await clickAway();

        expect(await screen.queryByText('Input is not a valid date.')).not.toBeInTheDocument();
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
        renderDatePicker({ value: JAN_1.toISOString() });

        expect(screen.getByRole('textbox')).toHaveValue('2025-01-01');
    });

    it('updates value when a valid date typed', async () => {
        const { clickAway, onDateChangeMock, typeInput } = renderDatePicker();

        await typeInput('2025-01-01');
        await clickAway();

        expect(onDateChangeMock).toHaveBeenCalledWith('2025-01-01', true);
    });

    it('updates value when an invalid date typed', async () => {
        const { clickAway, onDateChangeMock, typeInput } = renderDatePicker();

        await typeInput('2025');
        await clickAway();

        expect(onDateChangeMock).toHaveBeenCalledWith('2025', false);
    });

    it('sets the calendar date to match the valid', async () => {
        const { openCalendar } = renderDatePicker({ value: JAN_1.toISOString() });

        await openCalendar();

        const calendarDay = screen.getByRole('gridcell', { selected: true });

        expect(calendarDay).toBeInTheDocument();
        expect(calendarDay).toHaveTextContent('1');
    });
});

describe('ManagedDatePicker - input', () => {
    it('clears value if input is empty', async () => {
        const { clickAway, onDateChangeMock, typeInput } = renderDatePicker();

        await typeInput('2025-01-01');
        await clickAway();

        expect(onDateChangeMock).toHaveBeenCalledWith('2025-01-01', true);

        await typeInput('{backspace}'.repeat(10));

        await clickAway();

        expect(onDateChangeMock).toHaveBeenCalledWith('', true);
    });
});
