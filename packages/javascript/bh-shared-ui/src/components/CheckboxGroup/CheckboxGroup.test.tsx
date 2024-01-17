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
import { act, render, screen } from '../../test-utils';
import CheckboxGroup, { CheckboxGroupProps } from './CheckboxGroup';

const checkboxOptions: CheckboxGroupProps = {
    options: [
        { name: 'test-option 1', label: 'Test Option 1' },
        { name: 'test-option 2', label: 'Test Option 2' },
        { name: 'test-option 3', label: 'Test Option 3' },
    ],
    groupTitle: 'Test Options',
    handleCheckboxFilter: vi.fn(),
};

describe('Checkbox group', () => {
    const { groupTitle, options, handleCheckboxFilter } = checkboxOptions;

    beforeEach(() => {
        render(<CheckboxGroup groupTitle={groupTitle} handleCheckboxFilter={handleCheckboxFilter} options={options} />);
    });

    it('should render all of the options as checkboxes', () => {
        const testOption1 = screen.getByText(options[0].label);
        const testOption2 = screen.getByText(options[1].label);
        const testOption3 = screen.getByText(options[2].label);

        expect(testOption1).toBeInTheDocument();
        expect(testOption2).toBeInTheDocument();
        expect(testOption3).toBeInTheDocument();
    });

    it('should call the checkbox filter function when clicking on a checkbox input', async () => {
        const user = userEvent.setup();

        await act(async () => {
            await user.click(screen.getByText(options[0].label));
            await user.click(screen.getByText(options[1].label));
            await user.click(screen.getByText(options[2].label));
        });

        expect(handleCheckboxFilter).toHaveBeenCalledTimes(options.length);
    });

    it('should display the checkbox group title', () => {
        expect(screen.queryByText(groupTitle)).toBeInTheDocument();
    });
});
