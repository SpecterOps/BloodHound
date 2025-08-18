import { render, screen } from '../../../test-utils';
import CypherSearchMessage from './CypherSearchMessage';

const testMessageText = 'this is a test message';
const testMessageState = {
    showMessage: true,
    message: testMessageText,
};

const testClearMessage = vi.fn();

describe('CypherSearchMessage', () => {
    it('should display a message', () => {
        render(<CypherSearchMessage messageState={testMessageState} clearMessage={testClearMessage} />);
        expect(screen.getByText(/this is a test message/i)).toBeInTheDocument();
    });
});
