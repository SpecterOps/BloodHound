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
import { AssetGroupTagTypeTier, ConfigurationKey } from 'js-client-library';
import { longWait, render, screen, within } from '../../../test-utils';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import SummaryCard from './SummaryCard';

// Mock icons
vi.mock('../../../components/AppIcon/Icons/LargeRightArrow', () => ({
    default: () => <div data-testid='large-right-arrow' />,
}));

// Mock route and navigation
vi.mock('../../../routes', async () => {
    const actual = await vi.importActual('../../../routes');
    return {
        ...actual,
        ROUTE_ZONE_MANAGEMENT_DETAILS: 'details',
    }
});

const mockNavigate = vi.fn();
vi.mock('../../../../utils', async () => {
    const actual = await vi.importActual('../../../../utils');
    return {
        ...actual,
        useAppNavigate: () => mockNavigate,
    };
});

const configResponse = {
    data: [
        {
            key: ConfigurationKey.Tiering,
            value: { multi_tier_analysis_enabled: true, tier_limit: 3, label_limit: 10 },
        },
    ],
};

const server = setupServer();

describe('SummaryCard', () => {
    const props = {
        title: 'Test Tier',
        type: AssetGroupTagTypeTier,
        selectorCount: 7,
        memberCount: 13,
        id: 99,
        analysisEnabled: true,
    };

    const user = userEvent.setup();

    beforeEach(() => {
        mockNavigate.mockClear();
    });
    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('renders the summary card with title, selector count, and member count', async () => {
        render(<SummaryCard {...props} />);

        expect(await screen.findByText('Test Tier')).toBeInTheDocument();
        expect(screen.getByText('Selectors')).toBeInTheDocument();
        expect(screen.getByText('7')).toBeInTheDocument();
        expect(screen.getByText('Members')).toBeInTheDocument();
        expect(screen.getByText('13')).toBeInTheDocument();
    });

    it('renders icons', async () => {
        render(<SummaryCard {...props} />);
        expect(await screen.findAllByTestId('large-right-arrow')).toHaveLength(2);
    });

    it('navigates to the details page when "View Details" is clicked', async () => {
        render(<SummaryCard {...props} />);

        await user.click(await screen.findByText('View Details'));

        await longWait(() => {
            expect(mockNavigate).toHaveBeenCalledWith('/zone-management/details/tier/99');
        })
    });

    it('does not navigate when clicking other parts of the card', async () => {
        render(<SummaryCard {...props} />);
        await user.click(await screen.findByText('Test Tier'));

        expect(mockNavigate).not.toHaveBeenCalled();
    });

    it('does not render icon when multi tier analysis is disabled', async () => {
        const configRes = {
            data: [
                {
                    key: ConfigurationKey.Tiering,
                    value: { multi_tier_analysis_enabled: false, tier_limit: 3, label_limit: 10 },
                },
            ],
        };

        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configRes));
            }),
        );

        render(<SummaryCard {...props} />)

        const listItem = await screen.findByTestId('zone-management_summary_test_tier-list_item-99');
        expect(listItem).toBeInTheDocument();

        const icon = within(listItem).queryByTestId('analysis_disabled_icon');
        expect(icon).not.toBeInTheDocument();
    });

    it('renders tier icon tooltip when multi tier analysis is enabled but tier analysis is off', async () => {
        const props = {
            title: 'Test Tier',
            type: AssetGroupTagTypeTier,
            selectorCount: 7,
            memberCount: 13,
            id: 3,
            analysisEnabled: false,
        };

        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configResponse));
            }),
        );

        render(<SummaryCard {...props} />);

        const listItem = await screen.findByTestId('zone-management_summary_test_tier-list_item-3');
        expect(listItem).toBeInTheDocument();

        const icon = within(listItem).getByTestId('analysis_disabled_icon');
        expect(icon).toBeInTheDocument();
    });

    it('does not render tier icon tooltip when multi tier analysis is enabled and tier analysis is on', async () => {
        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configResponse));
            }),
        );

        render(<SummaryCard {...props} />);

        const listItem = await screen.findByTestId('zone-management_summary_test_tier-list_item-99');
        expect(listItem).toBeInTheDocument();

        const icon = within(listItem).queryByTestId('analysis_disabled_icon');
        expect(icon).not.toBeInTheDocument();
    });
});
