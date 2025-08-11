import { render, screen } from '../test-utils';
import { StatusIndicator } from './StatusIndicator';

describe('StatusIndicator', () => {
    it('displays a status indicator', () => {
        const { container } = render(<StatusIndicator status='good' />);
        expect(container.querySelector('circle')).toHaveClass('fill-[#BCD3A8]');
    });

    it('displays a status label', async () => {
        const { container } = render(<StatusIndicator status='bad' label='Bad' />);
        expect(container.querySelector('circle')).toHaveClass('fill-[#D9442E]');
        expect(container.querySelector('div span')?.childNodes).toHaveLength(2);
        const badStatus = await screen.findByText('Bad', {});
        expect(badStatus).toBeInTheDocument();
    });

    it('renders without label when label is empty string', () => {
        const { container } = render(<StatusIndicator status='good' label='' />);
        expect(container.querySelector('div span')?.childNodes).toHaveLength(1);
    });

    it('displays a pulsing indicator', async () => {
        const { container } = render(<StatusIndicator status='pending' pulse />);
        expect(container.querySelector('circle')).toHaveClass('fill-[#5CC3AD] animate-pulse');
    });

    it('renders without pulse animation by default', () => {
        const { container } = render(<StatusIndicator status='pending' />);
        expect(container.querySelector('circle')).not.toHaveClass('animate-pulse');
    });
});
