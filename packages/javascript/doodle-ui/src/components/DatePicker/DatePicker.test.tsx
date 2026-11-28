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

import { composeStories } from '@storybook/react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DateTime } from 'luxon';
import { expect } from 'vitest';
import * as stories from './DatePicker.stories';

const { Default, InputVariantUnderlined, BasicSingle } = composeStories(stories);

describe('DatePicker stories', () => {
    it.each([
        ['Default', Default],
        ['Underlined', InputVariantUnderlined],
        ['BasicSingle', BasicSingle],
    ] as const)('updates the %s input when a date is selected', async (_name, Story) => {
        const user = userEvent.setup();
        render(<Story />);

        await user.click(screen.getByRole('button', { name: 'Choose Date' }));
        await user.click(screen.getByRole('button', { name: 'Go to previous month' }));
        await user.click(screen.getByRole('gridcell', { name: '15' }));

        expect(screen.getByRole('textbox')).toHaveValue(
            DateTime.local().minus({ months: 1 }).set({ day: 15 }).toFormat('yyyy-MM-dd')
        );
    });

    it('reflects a typed date in the calendar', async () => {
        const user = userEvent.setup();
        render(<Default />);

        await user.type(screen.getByRole('textbox'), DateTime.local().set({ day: 15 }).toFormat('yyyy-MM-dd'));
        await user.click(screen.getByRole('button', { name: 'Choose Date' }));

        expect(screen.getByRole('gridcell', { name: '15' })).toHaveAttribute('aria-selected', 'true');
    });
});
