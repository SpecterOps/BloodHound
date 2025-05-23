import { render, screen } from '../../test-utils';
import GenericErrorBoundaryFallback from './GenericErrorBoundaryFallback';

describe('GenericErrorBoundaryFallack', () => {
    beforeEach(async () => {
        render(<GenericErrorBoundaryFallback />);
    });

    it('should display', () => {
        const elem = screen.getByRole('alert');

        expect(elem).toBeInTheDocument();
        expect(elem).toHaveTextContent('An unexpected error has occurred.');
        expect(screen.getByTestId('ErrorOutlineIcon')).toBeInTheDocument();
    });

    it('should be aligned to right of screen', () => {
        const elem = screen.getByRole('alert');
        const styles = getComputedStyle(elem);
        expect(styles.justifySelf).toEqual('flex-end');
    });
});
