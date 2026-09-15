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

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Tooltip } from './Tooltip';

describe('Tooltip', () => {
    it('renders above dialogs', async () => {
        const user = userEvent.setup();
        render(
            <Tooltip tooltip='Helpful context'>
                <button>Show tooltip</button>
            </Tooltip>
        );

        await user.hover(screen.getByRole('button', { name: 'Show tooltip' }));

        expect((await screen.findByRole('tooltip')).parentElement?.classList.contains('z-tooltip')).toBe(true);
    });
});
