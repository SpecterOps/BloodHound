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

import { render, screen } from '../../test-utils';
import GraphButton from './GraphButton';
import userEvent from '@testing-library/user-event';

const onClick = vi.fn();
const displayText = 'test';

describe('GraphButton', () => {
    it('should render a button with the passed in display text', () => {
        render(<GraphButton onClick={onClick} displayText={displayText} />);

        expect(screen.getByText(displayText)).toBeInTheDocument();
    });

    it('should call the onClick function when clicked', async () => {
        const user = userEvent.setup();
        render(<GraphButton onClick={onClick} displayText={displayText} />);

        await user.click(screen.getByRole('button'));

        expect(onClick).toHaveBeenCalled();
    });
});
