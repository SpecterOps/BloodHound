import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ListItemActionMenu from './ListItemActionMenu';

describe('ListItemActionMenu', () => {
    const testDeleteHandler = vitest.fn();
    const user = userEvent.setup();
    it('renders a ListItemActionMenu component', async () => {
        render(<ListItemActionMenu id={1} deleteQuery={testDeleteHandler} />);

        expect(screen.getByTestId('saved-query-action-menu-trigger')).toBeInTheDocument();
    });

    it('renders the popup content with run, edit/share, and delete when the menu trigger', async () => {
        render(<ListItemActionMenu id={1} deleteQuery={testDeleteHandler} />);

        expect(screen.getByTestId('saved-query-action-menu-trigger')).toBeInTheDocument();

        await user.click(screen.getByRole('button'));
        expect(screen.getByText(/run/i)).toBeInTheDocument();
        expect(screen.getByText(/edit\/share/i)).toBeInTheDocument();
        expect(screen.getByText(/delete/i)).toBeInTheDocument();
    });

    it('fires delete when edit is clicked', async () => {
        render(<ListItemActionMenu id={1} deleteQuery={testDeleteHandler} />);

        expect(screen.getByTestId('saved-query-action-menu-trigger')).toBeInTheDocument();

        await user.click(screen.getByRole('button'));

        await user.click(screen.getByText(/delete/i));
        expect(testDeleteHandler).toBeCalled();
    });

    it('closes', async () => {
        render(<ListItemActionMenu id={1} deleteQuery={testDeleteHandler} />);

        expect(screen.getByTestId('saved-query-action-menu-trigger')).toBeInTheDocument();

        await user.click(screen.getByRole('button'));
        expect(screen.getByText(/edit\/share/i)).toBeInTheDocument();

        await user.click(screen.getByRole('button'));
        expect(screen.queryByText(/edit\/share/i)).not.toBeInTheDocument();
    });
});
