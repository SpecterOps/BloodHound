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
import { AssetGroupTagTypeTier } from 'js-client-library';
import { render, screen } from '../../../test-utils';
import SummaryCard from './SummaryCard';

// Mock icons
vi.mock('../../../components/AppIcon/Icons/LargeRightArrow', () => ({
    default: () => <div data-testid='large-right-arrow' />,
}));

// Mock route and navigation
vi.mock('../../../routes', () => ({
    ROUTE_TIER_MANAGEMENT_DETAILS: 'details',
}));

const mockNavigate = vi.fn();
vi.mock('../../../utils', () => ({
    useAppNavigate: () => mockNavigate,
}));

describe('SummaryCard', () => {
    const props = {
        title: 'Test Tier',
        type: AssetGroupTagTypeTier,
        selectorCount: 7,
        memberCount: 13,
        id: 99,
    };

    const user = userEvent.setup();

    beforeEach(() => {
        mockNavigate.mockClear();
    });

    it('renders the summary card with title, selector count, and member count', () => {
        render(<SummaryCard {...props} />);

        expect(screen.getByText('Test Tier')).toBeInTheDocument();
        expect(screen.getByText('Selectors')).toBeInTheDocument();
        expect(screen.getByText('7')).toBeInTheDocument();
        expect(screen.getByText('Members')).toBeInTheDocument();
        expect(screen.getByText('13')).toBeInTheDocument();
    });

    it('renders icons', () => {
        render(<SummaryCard {...props} />);
        expect(screen.getAllByTestId('large-right-arrow')).toHaveLength(2);
    });

    it('navigates to the details page when "View Details" is clicked', async () => {
        render(<SummaryCard {...props} />);
        await user.click(screen.getByText('View Details'));

        expect(mockNavigate).toHaveBeenCalledWith('/tier-management/details/tier/99');
    });

    it('does not navigate when clicking other parts of the card', async () => {
        render(<SummaryCard {...props} />);
        await user.click(screen.getByText('Test Tier'));

        expect(mockNavigate).not.toHaveBeenCalled();
    });
});
