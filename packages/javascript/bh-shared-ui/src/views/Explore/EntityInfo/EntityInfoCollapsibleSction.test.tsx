import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../../../test-utils';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';

describe('EntityInfoCollapsibleSection', () => {
    it('renders an error message without throwing a TypeError', async () => {
        const user = userEvent.setup();
        const testLabel = 'Section';
        const testCount = 100;
        const testOnChange = vi.fn();
        const testIsLoading = false;
        const testIsError = true;
        const isExpanded = true;
        const error = {};

        render(
            <EntityInfoCollapsibleSection
                label={testLabel}
                count={testCount}
                onChange={testOnChange}
                isLoading={testIsLoading}
                isError={testIsError}
                error={error}
                isExpanded={isExpanded}
            />
        );

        expect(screen.getByText(testLabel)).toBeInTheDocument();
        user.click(screen.getByText(testLabel));
        await waitFor(() =>
            expect(screen.getByText('An unknown error occurred during the request.')).toBeInTheDocument()
        );
    });
});
