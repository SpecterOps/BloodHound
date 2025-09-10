import { faGem, faVolcano } from '@fortawesome/free-solid-svg-icons';
import userEvent from '@testing-library/user-event';
import { act, render } from '../../test-utils';
import DropdownSelector from './DropdownSelector';

const testDropdownOptions = [
    { key: 1, value: 'Test 1', icon: faGem },
    { key: 2, value: 'Test 2', icon: faVolcano },
    { key: 3, value: 'Test 3' },
];

const onChange = vi.fn();

describe('DropdownSelector', () => {
    it('renders a primary button as expected', async () => {
        const user = userEvent.setup();
        const screen = await act(async () => {
            return render(<DropdownSelector options={testDropdownOptions} selectedText='Test' onChange={onChange} />);
        });
        const button = await screen.getByRole('button');
        expect(button).toHaveClass('bg-primary');
        expect(button).toHaveClass('rounded-3xl');
        expect(button).toHaveClass('uppercase');
        await user.click(button);
        const listItems = screen.getAllByRole('listitem');
        expect(listItems).toHaveLength(3);
    });
    it('renders icons as expected', async () => {
        const user = userEvent.setup();
        const screen = await act(async () => {
            return render(<DropdownSelector options={testDropdownOptions} selectedText='Test' onChange={onChange} />);
        });

        await user.click(screen.getByRole('button'));
        expect(await screen.findByText('gem')).toBeInTheDocument();
        expect(await screen.findByText('volcano')).toBeInTheDocument();
    });
    it('renders a non-primary button as expected', async () => {
        const user = userEvent.setup();
        const screen = await act(async () => {
            return render(
                <DropdownSelector
                    options={testDropdownOptions}
                    variant='transparent'
                    selectedText='Test'
                    onChange={onChange}
                />
            );
        });

        expect(await screen.findByText('Test')).toBeInTheDocument();
        const button = await screen.getByRole('button');
        expect(button).toHaveClass('rounded-md');
        expect(button).toHaveClass('bg-transparent');
        await user.click(button);
        expect(await screen.findByText('Test 1')).toBeInTheDocument();
        expect(await screen.findByText('Test 2')).toBeInTheDocument();
        expect(await screen.findByText('Test 3')).toBeInTheDocument();
    });
    it('handles a selection as expected', async () => {
        const user = userEvent.setup();
        const screen = await act(async () => {
            return render(<DropdownSelector options={testDropdownOptions} selectedText='Test' onChange={onChange} />);
        });

        const button = await screen.getByRole('button');
        await user.click(button);
        const selection = await screen.findByText('Test 1');
        await user.click(selection);
        expect(onChange).toHaveBeenCalled();
    });
});
