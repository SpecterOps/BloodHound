import { render, screen } from '../test-utils';
import { StatusIndicator } from './StatusIndicator';

describe('StatusIndicator', () => {
    it('displays a status indicator', () => {
        const { container } = render(<StatusIndicator type='good' />);
        expect(container.querySelector('circle')).toHaveClass('fill-[#BCD3A8]');
    });

    it('displays a status label', async () => {
        const { container } = render(<StatusIndicator type='bad' label='Bad' />);
        expect(container.querySelector('circle')).toHaveClass('fill-[#D9442E]');
        const badStatus = await screen.findByText('Bad', {});
        expect(badStatus).toBeInTheDocument();
    });

    it('displays a pulsing indicator', async () => {
        const { container } = render(<StatusIndicator type='pending' pulse />);
        expect(container.querySelector('circle')).toHaveClass('fill-[#5CC3AD] animate-pulse');
    });
});
