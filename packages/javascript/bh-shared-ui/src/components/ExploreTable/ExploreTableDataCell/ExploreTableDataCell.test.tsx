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
import ExploreTableDataCell from '.';
import { fireEvent, render } from '../../../test-utils';

describe('ExploreTableDataCell', () => {
    const cellValue = '123';
    const setup = () => {
        const screen = render(<ExploreTableDataCell value={cellValue} columnKey={'objectId'} />);
        const user = userEvent.setup();
        return { screen, user };
    };
    it('should copy the cells text if the cell is a string, or array of strings', async () => {
        const { screen, user } = setup();

        const copyButton = screen.getByText('copy');
        await user.click(copyButton);

        const clipBoard = await navigator.clipboard.readText();
        expect(clipBoard).toBe(cellValue);
    });
    it('temporarily displays the checkmark icon after clicking copy', async () => {
        const { screen, user } = setup();

        await user.click(screen.getByText('copy'));

        const checkmark = await screen.findByText('check');
        expect(checkmark).toBeInTheDocument();

        const animatedElement = screen.getByRole('button');

        animatedElement.role = 'button'; // Looks like JSDOM does not apply the role property implicitly?
        fireEvent.animationEnd(animatedElement); // Manually trigger the animationend event
        const copyButton = await screen.findByText('copy');
        expect(copyButton).toBeInTheDocument();
    });
});
