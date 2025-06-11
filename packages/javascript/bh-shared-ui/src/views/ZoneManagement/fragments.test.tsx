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
import { render, screen, waitFor } from '../../test-utils';
import { TierActionBar } from './fragments';

vi.mock('../../utils', async () => {
    const actual = await vi.importActual<typeof import('../../utils')>('../../utils');
    return {
        ...actual,
        useAppNavigate: () => mockNavigate,
    };
});

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
    return {
        ...actual,
        Link: ({ to, children }: { to: string; children: React.ReactNode }) => <a href={to}>{children}</a>,
    };
});

const mockNavigate = vi.fn();

describe('TierActionBar', () => {
    beforeEach(() => {
        vi.stubGlobal('location', {
            pathname: '/zone-management/summary',
        } as unknown as Location);
    });

    it('renders Tier create button', () => {
        render(<TierActionBar tierId='1' labelId={undefined} selectorId={undefined} />);

        expect(screen.getByText('Create Tier')).toBeInTheDocument();
    });

    it('disables CreateMenus when tierId or labelId is missing', () => {
        render(<TierActionBar tierId='' labelId={undefined} selectorId={undefined} />);

        expect(screen.getByText('Create Tier').closest('button')).toBeDisabled();
    });

    it('renders Create Label button and navigates when labelId is set', async () => {
        vi.stubGlobal('location', {
            pathname: '/zone-management/summary/label/10',
        } as unknown as Location);

        render(<TierActionBar tierId='1' labelId='10' selectorId={undefined} />);

        expect(screen.getByText('Create Label')).toBeInTheDocument();

        await userEvent.click(screen.getByText('Create Label'));
        await waitFor(() => {
            expect(mockNavigate).toHaveBeenCalledWith('/zone-management/save/label/10');
        });
    });

    it('shows the Edit button when showEditButton and getSavePath are provided', () => {
        render(
            <TierActionBar
                tierId='1'
                labelId='10'
                selectorId='3'
                showEditButton={true}
                getSavePath={(t, l, s) => `/edit/path/${t}/${l}/${s}`}
            />
        );

        const editLink = screen.getByText('Edit') as HTMLAnchorElement;
        expect(editLink).toBeInTheDocument();
        expect(editLink.getAttribute('href')).toBe('/edit/path/1/10/3');
    });
});
