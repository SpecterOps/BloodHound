import userEvent from '@testing-library/user-event';

import { render, screen } from '../../test-utils';
import { DetailsAccordion } from './DetailsAccordion';

// Need test ID for icon
vi.mock('@fortawesome/react-fontawesome', () => ({
    FontAwesomeIcon: (props: any) => <span data-testid='fa' {...props} />,
}));

vi.mock('@fortawesome/free-solid-svg-icons', () => ({ faAngleDown: {} }));

type Item = { title: string; description: string; disabled?: boolean };

const Header: React.FC<Item> = (item) => <h3>{item.title}</h3>;
const Content: React.FC<Item> = (item) => <p>{item.description}</p>;
const Empty: React.FC = () => <div role='status'>No items</div>;

describe('<DetailsAccordion />', () => {
    it('renders Empty when items is undefined and Empty is provided', () => {
        render(<DetailsAccordion<Item> Header={Header} Content={Content} Empty={Empty} />);
        expect(screen.getByRole('status')).toHaveTextContent('No items');
    });

    it('renders a single item when items is an object', () => {
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={{ title: 'One', description: 'Only child' }}
            />
        );
        expect(screen.getByRole('button')).toHaveTextContent('One');
    });

    it('passes item props to Header and Content', async () => {
        const user = userEvent.setup();
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[
                    { title: 'A', description: 'Alpha' },
                    { title: 'B', description: 'Beta' },
                ]}
                openIndex={0}
            />
        );

        // default open is index 0 -> content "Alpha" should be visible
        expect(screen.getByText('Alpha')).toBeInTheDocument();

        // Click header B, then "Beta" should appear
        await user.click(screen.getAllByRole('button').find((el) => el.textContent?.includes('B'))!);
        expect(screen.getByText('Beta')).toBeInTheDocument();
    });

    it('calls getKey for each item (stable keys hook)', () => {
        const items = [
            { title: 'A', description: 'Alpha' },
            { title: 'B', description: 'Beta' },
        ];
        const getKey = vi.fn((i: Item) => i.title);

        render(<DetailsAccordion<Item> Header={Header} Content={Content} items={items} getKey={getKey} />);

        expect(getKey).toHaveBeenCalledTimes(2);
        expect(getKey).toHaveBeenNthCalledWith(1, items[0], 0);
        expect(getKey).toHaveBeenNthCalledWith(2, items[1], 1);
    });

    it('skips undefined items gracefully', () => {
        render(
            <DetailsAccordion<any>
                Header={({ label }: any) => <div>{label}</div>}
                Content={() => <div>ok</div>}
                // an undefined item
                items={[{ label: 'Valid' }, undefined]}
            />
        );
        expect(screen.getByRole('button')).toHaveTextContent('Valid');
        // only one header rendered
        expect(screen.getAllByRole('button')).toHaveLength(1);
    });

    it('applies accent styling to headers when accent is true', () => {
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[{ title: 'Fancy', description: 'With accent' }]}
                accent
            />
        );

        const header = screen.getByRole('button');
        // spot-check a couple of classes from implementation
        expect(header).toHaveClass('bg-[#e0e0e0]');
        expect(header).toHaveClass('border-l-8');
        expect(header).toHaveClass('border-primary');
    });

    it('disables items via itemDisabled and hides the chevron icon visually', () => {
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[
                    { title: 'Disabled', description: 'Nope', disabled: true },
                    { title: 'Enabled', description: 'Yep', disabled: false },
                ]}
                itemDisabled={(i) => !!i.disabled}
            />
        );

        const [disabledHeader, enabledHeader] = screen.getAllByRole('button');

        // Icon for disabled row should have opacity-0
        const disabledIcon = disabledHeader.querySelector('[data-testid="fa"]');
        expect(disabledIcon).toHaveClass('opacity-0');

        const enabledIcon = enabledHeader.querySelector('[data-testid="fa"]');
        expect(enabledIcon).not.toHaveClass('opacity-0');
    });

    it('respects openIndex (default opened panel)', () => {
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[
                    { title: 'First', description: 'one' },
                    { title: 'Second', description: 'two' },
                ]}
                openIndex={1}
            />
        );
        // only 'two' should be visible initially
        expect(screen.queryByText('one')).not.toBeInTheDocument();
        expect(screen.getByText('two')).toBeInTheDocument();
    });

    it('clicking a disabled header does not toggle content', async () => {
        const user = userEvent.setup();
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[{ title: 'Locked', description: 'hidden', disabled: true }]}
                itemDisabled={(i) => !!i.disabled}
            />
        );

        const header = screen.getByRole('button', { name: /Locked/i });
        await user.click(header);

        // never opens
        expect(screen.queryByText('hidden')).not.toBeInTheDocument();
    });

    it('clicking an enabled header toggles content (single mode)', async () => {
        const user = userEvent.setup();
        render(
            <DetailsAccordion<Item>
                Header={Header}
                Content={Content}
                items={[
                    { title: 'A', description: 'Alpha' },
                    { title: 'B', description: 'Beta' },
                ]}
            />
        );

        // Initially nothing is open (no openIndex)
        expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
        expect(screen.queryByText('Beta')).not.toBeInTheDocument();

        // Open A
        await user.click(screen.getByRole('button', { name: /A/i }));
        expect(screen.getByText('Alpha')).toBeInTheDocument();

        // Open B (single mode should close A)
        await user.click(screen.getByRole('button', { name: /B/i }));
        expect(screen.queryByText('Alpha')).not.toBeInTheDocument();
        expect(screen.getByText('Beta')).toBeInTheDocument();
    });
});
