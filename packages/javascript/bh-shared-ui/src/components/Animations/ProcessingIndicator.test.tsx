import { render, screen } from '../../test-utils';
import ProcessingIndicator from './ProcessingIndicator';

describe('ProcessingIndicator', () => {
    it('renders the title', () => {
        render(<ProcessingIndicator title='Analyzing' />);
        expect(screen.getByText('Analyzing')).toBeInTheDocument();
    });

    it('renders three animated dots', () => {
        render(<ProcessingIndicator title='Loading' />);

        const dots = screen.getAllByText('.');
        expect(dots).toHaveLength(3);

        // Check animation classes
        dots.forEach((dot) => {
            expect(dot).toHaveClass('animate-pulse');
        });

        // Check animation delays
        expect(dots[0]).toHaveStyle({ 'animation-delay': undefined }); // no delay set
        expect(dots[1]).toHaveStyle({ 'animation-delay': '0.2s' });
        expect(dots[2]).toHaveStyle({ 'animation-delay': '0.4s' });
    });

    it('applies animation class to the title', () => {
        render(<ProcessingIndicator title='Analyzing' />);
        const title = screen.getByText('Analyzing');
        expect(title).toHaveClass('animate-pulse');
    });
});
