import { render, screen } from '@testing-library/react';
import { useQuery } from 'react-query';
import { vi } from 'vitest';
import ObjectCountPanel from './ObjectCountPanel';

vi.mock('react-query', () => ({
    useQuery: vi.fn(),
}));

vi.mock('../../../utils', () => ({
    apiClient: {
        getAssetGroupMembersCount: vi.fn(),
    },
}));

describe('ObjectCountPanel', () => {
    it('renders error message on error', () => {
        (useQuery as jest.Mock).mockReturnValue({ isError: true });
        render(<ObjectCountPanel selectedTier={1} />);

        expect(screen.getByText('There was an error fetching this data')).toBeInTheDocument();
    });

    it('renders the total count and object counts on success', () => {
        (useQuery as jest.Mock).mockReturnValue({
            isSuccess: true,
            data: {
                total_count: 100,
                counts: { 'Object A': 50, 'Object B': 30, 'Object C': 20 },
            },
        });

        render(<ObjectCountPanel selectedTier={1} />);

        expect(screen.getByText('Total Count')).toBeInTheDocument();
        expect(screen.getByText('100')).toBeInTheDocument();
        expect(screen.getByText('Object A')).toBeInTheDocument();
        expect(screen.getByText('50')).toBeInTheDocument();
        expect(screen.getByText('Object B')).toBeInTheDocument();
        expect(screen.getByText('30')).toBeInTheDocument();
        expect(screen.getByText('Object C')).toBeInTheDocument();
        expect(screen.getByText('20')).toBeInTheDocument();
    });
});
