// Test file to verify SniffDeepSearch component functionality
// Copyright 2025 Specter Ops, Inc.

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { BrowserRouter } from 'react-router-dom';
import { usePathfindingFilters, usePathfindingSearch } from 'bh-shared-ui';
import { act } from '../../../test-utils';
import SniffDeepSearch from './SniffDeepSearch';

describe('SniffDeepSearch', () => {
    const server = setupServer(
        rest.get('/api/v2/search', (req, res, ctx) => {
            return res(ctx.json({ data: [] }));
        })
    );

    const MockSniffDeepSearchWrapper = () => {
        const pathfindingSearchState = usePathfindingSearch();
        const pathfindingFilterState = usePathfindingFilters();
        
        return (
            <BrowserRouter>
                <SniffDeepSearch
                    pathfindingSearchState={pathfindingSearchState}
                    pathfindingFilterState={pathfindingFilterState}
                />
            </BrowserRouter>
        );
    };

    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('renders the dropdown with All and DCSync options', async () => {
        await act(async () => {
            render(<MockSniffDeepSearchWrapper />);
        });

        // Check that the dropdown button is rendered with "All" as default
        expect(screen.getByRole('button', { name: /all/i })).toBeInTheDocument();
    });

    it('allows user to change dropdown selection from All to DCSync', async () => {
        const user = userEvent.setup();
        
        await act(async () => {
            render(<MockSniffDeepSearchWrapper />);
        });

        const dropdownButton = screen.getByRole('button', { name: /all/i });
        await user.click(dropdownButton);

        // Check that both options are visible
        expect(screen.getByText('All')).toBeInTheDocument();
        expect(screen.getByText('DCSync')).toBeInTheDocument();

        // Click DCSync option
        await user.click(screen.getByText('DCSync'));

        // The button should now show DCSync
        expect(screen.getByRole('button', { name: /dcsync/i })).toBeInTheDocument();
    });

    it('renders pathfinding search components below the dropdown', async () => {
        await act(async () => {
            render(<MockSniffDeepSearchWrapper />);
        });

        // Check that the pathfinding search elements are present
        expect(screen.getByLabelText(/start node/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/destination node/i)).toBeInTheDocument();
    });
});
