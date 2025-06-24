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

import { SeedTypeCypher } from 'js-client-library';
import DeleteSelectorButton from '.';
import { render, screen } from '../../../../../test-utils';

const testSelector = {
    id: 777,
    asset_group_tag_id: 1,
    name: 'foo',
    allow_disable: true,
    description: 'bar',
    is_default: true,
    auto_certify: true,
    created_at: '2024-10-05T17:54:32.245Z',
    created_by: 'Stephen64@gmail.com',
    updated_at: '2024-07-20T11:22:18.219Z',
    updated_by: 'Donna13@yahoo.com',
    disabled_at: '2024-09-15T09:55:04.177Z',
    disabled_by: 'Roberta_Morar72@hotmail.com',
    count: 3821,
    seeds: [{ selector_id: 777, type: SeedTypeCypher, value: 'match(n) return n limit 5' }],
};

describe('Delete Selector Button rendering', () => {
    it('does not render when selectorId is blank', () => {
        render(<DeleteSelectorButton selectorId='' selectorData={undefined} onClick={vi.fn} />);

        expect(screen.queryByRole('button', { name: /Delete Selector/ })).not.toBeInTheDocument();
    });

    it('does not render when selector data is not defined', () => {
        render(<DeleteSelectorButton selectorId='1' selectorData={undefined} onClick={vi.fn} />);

        expect(screen.queryByRole('button', { name: /Delete Selector/ })).not.toBeInTheDocument();
    });

    it('does not render when selector is default', () => {
        render(<DeleteSelectorButton selectorId='1' selectorData={testSelector} onClick={vi.fn} />);

        expect(screen.queryByRole('button', { name: /Delete Selector/ })).not.toBeInTheDocument();
    });

    it('renders when all data is available and selector is not default', () => {
        render(
            <DeleteSelectorButton
                selectorId='1'
                selectorData={{ ...testSelector, is_default: false }}
                onClick={vi.fn}
            />
        );

        expect(screen.getByRole('button', { name: /Delete Selector/ })).toBeInTheDocument();
    });
});
