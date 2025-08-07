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
import { DatePicker } from './DatePicker';

describe('DatePicker', () => {
    const values = {
        label: 'testLabel',
        onSelect: vi.fn(),
    };

    beforeEach(() => {
        render(<DatePicker label={values.label} onSelect={values.onSelect} setValue={vi.fn()} />);
    });

    it('renders a button for selecting a date', () => {
        const input = screen.getByRole('textbox');
        expect(input).toBeInTheDocument();
    });

    it('renders a label', () => {
        const label = screen.getByText(values.label);
        expect(label).toBeInTheDocument();
    });

    it('renders placeholder text', () => {
        const placeholder = screen.queryByPlaceholderText(/yyyy-mm-dd/);
        expect(placeholder).toBeInTheDocument();
    });

    it('opens a calendar picker on click', async () => {
        const user = userEvent.setup();
        const button = screen.getByText(/calendar-day/);

        await user.click(button);

        const calendar = screen.getByRole('grid');
        expect(calendar).toBeInTheDocument();
    });
});
