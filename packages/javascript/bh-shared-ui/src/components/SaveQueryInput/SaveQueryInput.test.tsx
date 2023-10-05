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
import { apiClient } from '../..';
import SaveQueryInput from './SaveQueryInput';

describe('SaveQueryInput', () => {
    beforeEach(() => {
        render(<SaveQueryInput cypherQuery='test query' />);
    });

    it('should open the text input when the save button is clicked', async () => {
        const user = userEvent.setup();

        const saveButton = screen.getByRole('button', { name: /floppy-disk/i });
        expect(saveButton).toBeInTheDocument();

        expect(screen.queryByRole('textbox', { name: /search name/i })).not.toBeInTheDocument();

        await user.click(saveButton);

        const textInput = screen.getByRole('textbox', { name: /search name/i });
        expect(textInput).toBeInTheDocument();
    });

    it('should disable the save button when no name has been provided to the text input', async () => {
        const user = userEvent.setup();

        const saveButton = screen.getByRole('button', { name: /floppy-disk/i });
        expect(saveButton).toBeEnabled();

        await user.click(saveButton);

        expect(saveButton).toBeDisabled();
    });

    it('should handle user input', async () => {
        const apiSpy = vi.spyOn(apiClient, 'createUserQuery');
        // @ts-ignore mock the apiClient instead of msw.setupServer(), we dont use the apiClient's response within the component
        apiSpy.mockReturnValue({ dont: 'care' });

        const user = userEvent.setup();

        const saveButton = screen.getByRole('button', { name: /floppy-disk/i });
        expect(saveButton).toBeEnabled();

        // open text input
        await user.click(saveButton);

        const textInput = screen.getByRole('textbox', { name: /search name/i });
        await user.type(textInput, 'my favorite cypher query');

        await user.click(saveButton);

        expect(apiSpy).toHaveBeenCalledTimes(1);
        expect(apiSpy).toHaveBeenCalledWith({ name: 'my favorite cypher query', query: 'test query' });
    });
});
