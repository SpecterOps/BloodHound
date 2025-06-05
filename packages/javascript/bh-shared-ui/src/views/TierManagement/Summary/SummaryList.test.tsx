// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// http://www.apache.org/licenses/LICENSE-2.0
//
// SPDX-License-Identifier: Apache-2.0

import userEvent from '@testing-library/user-event';
import { AssetGroupTagsListItem } from 'js-client-library';
import { UseQueryResult } from 'react-query';
import { render, screen } from '../../../test-utils';
import SummaryList from './SummaryList';

vi.mock('../../../components/AppIcon/Icons/DownArrow', () => ({
    default: () => <div data-testid='tier-management_summary-list_down-arrow' />,
}));

vi.mock('./SummaryCard', () => ({
    default: ({ title }: { title: string }) => <div data-testid='tier-management_summary-list_card'>{title}</div>,
}));

vi.mock('../Details/utils', () => ({
    itemSkeletons: [
        () => <li key='skeleton-1' data-testid='tier-management_summary-list_loading-skeleton' />,
        () => <li key='skeleton-2' data-testid='tier-management_summary-list_loading-skeleton' />,
        () => <li key='skeleton-3' data-testid='tier-management_summary-list_loading-skeleton' />,
    ],
}));

const mockData: AssetGroupTagsListItem[] = [
    {
        id: 1,
        name: 'Mock Tier 1',
        kind_id: 10,
        type: 1,
        position: 1,
        requireCertify: false,
        description: 'A mock tag for testing',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
        deleted_at: null,
        created_by: 'user',
        updated_by: 'user',
        deleted_by: null,
        counts: {
            selectors: 5,
            members: 10,
        },
    },
    {
        id: 2,
        name: 'Mock Tier 2',
        kind_id: 20,
        type: 1,
        position: 2,
        requireCertify: true,
        description: 'Another mock tag',
        created_at: '2024-01-03T00:00:00Z',
        updated_at: '2024-01-04T00:00:00Z',
        deleted_at: null,
        created_by: 'user',
        updated_by: 'user',
        deleted_by: null,
        counts: {
            selectors: 3,
            members: 6,
        },
    },
];

const expectedCount = mockData.filter((item) => item.type === 1).length;

describe('SummaryList', () => {
    it('shows skeletons when loading', () => {
        const query = {
            isLoading: true,
            isError: false,
        } as unknown as UseQueryResult<AssetGroupTagsListItem[]>;

        render(<SummaryList title='Tiers' selected='' listQuery={query} onSelect={() => {}} />);

        expect(screen.getAllByTestId('tier-management_tiers-list_loading-skeleton')).toHaveLength(3);
    });

    it('shows an error message when query fails', async () => {
        const query = {
            isLoading: false,
            isError: true,
        } as unknown as UseQueryResult<AssetGroupTagsListItem[]>;

        render(<SummaryList title='Tiers' selected='' listQuery={query} onSelect={() => {}} />);

        expect(await screen.findByText('There was an error fetching this data')).toBeInTheDocument();
    });

    it('renders sorted list items by name descending', async () => {
        const query = {
            isSuccess: true,
            data: mockData,
        } as unknown as UseQueryResult<AssetGroupTagsListItem[]>;

        render(<SummaryList title='Tiers' selected='' listQuery={query} onSelect={() => {}} />);

        const cards = await screen.findAllByTestId('tier-management_summary-list_card');
        expect(cards[0]).toHaveTextContent('Mock Tier 2');
        expect(cards[1]).toHaveTextContent('Mock Tier 1');
    });

    it('renders a down arrow only for items of type 1', async () => {
        const query = {
            isSuccess: true,
            data: mockData,
        } as unknown as UseQueryResult<AssetGroupTagsListItem[]>;

        render(<SummaryList title='Labels' selected='' listQuery={query} onSelect={() => {}} />);

        const arrows = await screen.findAllByTestId('tier-management_summary-list_down-arrow');
        expect(arrows).toHaveLength(expectedCount);
    });

    it('calls onSelect when a card is clicked', async () => {
        const onSelect = vi.fn();

        const query = {
            isSuccess: true,
            data: [mockData[0]],
        } as unknown as UseQueryResult<AssetGroupTagsListItem[]>;

        render(<SummaryList title='Tiers' selected='' listQuery={query} onSelect={onSelect} />);

        await userEvent.click(await screen.findByTestId('tier-management_summary-list_card'));
        expect(onSelect).toHaveBeenCalledWith(mockData[0].id);
    });

    it('applies highlight border for selected item', async () => {
        const query = {
            isSuccess: true,
            data: [mockData[0]],
        } as unknown as UseQueryResult<AssetGroupTagsListItem[]>;

        const { container } = render(
            <SummaryList title='Labels' selected={mockData[0].id.toString()} listQuery={query} onSelect={() => {}} />
        );

        const selectedItem = container.querySelector('li');
        expect(selectedItem?.className).toMatch(/border.*rounded-xl/);
    });
});
