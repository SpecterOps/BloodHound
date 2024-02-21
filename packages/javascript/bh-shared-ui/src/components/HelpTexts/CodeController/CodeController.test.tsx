// Copyright 2024 Specter Ops, Inc.
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
import { Screen, act, render } from '../../../test-utils';
import CodeController from './CodeController';

describe('CodeController', () => {
    const defaultExpected = 'testing some code to display';
    const setup = async (code = defaultExpected) => {
        const user = userEvent.setup();
        const screen: Screen = await act(async () => {
            return render(<CodeController>{code}</CodeController>);
        });
        return { user, screen };
    };

    it('displays the value thats passed via children', async () => {
        const { screen } = await setup();

        expect(screen.getByText(defaultExpected)).toBeInTheDocument();
    });

    it('defaults to wrapped and removes .wrapped class when unwrap btn is clicked', async () => {
        const { screen, user } = await setup();

        expect(screen.getByText(defaultExpected).className.includes('wrapped')).toBeTruthy();

        await user.click(screen.getByText('Unwrap'));

        expect(screen.getByText(defaultExpected).className.includes('wrapped')).toBeFalsy();
    });

    it('copys children value to clipboard after clicking the copy btn', async () => {
        const { screen, user } = await setup();

        await user.click(screen.getByText('copy'));

        const copiedText = await window.navigator.clipboard.readText();

        expect(copiedText).toBe(defaultExpected);
    });

    it('indicates the code container is scrollable when the code is unwrapped', async () => {
        const expected = Array(10).fill('testing large code block').join(' ');
        const { screen, user } = await setup(expected);

        await user.click(screen.getByText('Unwrap'));

        const codeContainer = screen.getByText(expected);

        expect(codeContainer.className.includes('scrollLeft')).toBeTruthy();
    });
});
