import { render, screen } from '@testing-library/react';
import SalesMessage from './SalesMessage';

const mockUseGetConfiguration = vi.fn();
vi.mock('../../hooks', async () => {
    const actual = await vi.importActual('../../hooks');
    return {
        ...actual,
        useGetConfiguration: () => mockUseGetConfiguration,
    };
});

const mockParseTieringConfiguration = vi.fn();
vi.mock('js-client-library', async () => {
    const actual = await vi.importActual('js-client-library');
    return {
        ...actual,
        parseTieringConfiguration: () => mockParseTieringConfiguration,
    };
});

describe('SalesMessage', async () => {
    it('renders sales message when multi_tier_analysis_enabled is false', () => {
        (mockUseGetConfiguration).mockReturnValue({
            data: [{
                key: "analysis.tiering",
                name: "Multi-Tier Analysis Configuration",
                value: {
                    label_limit: 10,
                    multi_tier_analysis_enabled: false,
                    tier_limit: 3
                }
            }],
        });

        (mockParseTieringConfiguration).mockReturnValue({
            value: { multi_tier_analysis_enabled: false },
        });

        render(<SalesMessage />);

        expect(
            screen.getByText(/Upgrade Privilege Zones/i)
        ).toBeInTheDocument();
    });

    it('does not render sales message when multi_tier_analysis_enabled is true', () => {
        (mockUseGetConfiguration).mockReturnValue({
            data: [{
                key: "analysis.tiering",
                name: "Multi-Tier Analysis Configuration",
                value: {
                    label_limit: 10,
                    multi_tier_analysis_enabled: true,
                    tier_limit: 3
                }
            }],
        });

        (mockParseTieringConfiguration).mockReturnValue({
            value: { multi_tier_analysis_enabled: true },
        });

        render(<SalesMessage />);

        expect(
            screen.queryByText(/Upgrade Privilege Zones/i)
        ).not.toBeInTheDocument();
    });
});
