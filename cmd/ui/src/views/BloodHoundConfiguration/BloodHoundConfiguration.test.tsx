import { render, screen } from 'src/test-utils';
import BloodHoundConfiguration from './BloodHoundConfiguration';

describe('BloodHoundConfiguration', () => {
    beforeEach(() => {
        render(<BloodHoundConfiguration />);
    });

    it('renders with a title', () => {
        const title = screen.getByRole('heading', { name: /BloodHound Configuration/i });
        expect(title).toBeInTheDocument();
    });

    it('renders a link to the documentation', () => {
        const link = screen.getByRole('link');
        expect(link).toHaveTextContent('documentation');
    });

    it('renders citrix config controls', () => {
        const title = screen.getByRole('heading', { name: /Citrix RDP Support/i });
        expect(title).toBeInTheDocument();
    });
});
