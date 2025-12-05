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
import { render, screen } from '../../../test-utils';
import { MembersList } from './MembersList';

describe('MembersList', () => {
    it('sorting the list fires the onChangeSortOrder callback', async () => {
        const user = userEvent.setup();
        const testOnChangeSortOrder = vi.fn();

        render(
            <MembersList
                listQuery={{} as any}
                selected='1'
                onClick={vi.fn()}
                onChangeSortOrder={testOnChangeSortOrder}
                sortOrder='asc'
            />
        );

        await user.click(screen.getByText('Objects', { exact: false }));

        expect(testOnChangeSortOrder).toBeCalledWith('desc');
    });
});
