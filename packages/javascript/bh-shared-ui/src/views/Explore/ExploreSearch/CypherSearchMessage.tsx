import { useRef } from 'react';
import { Transition } from 'react-transition-group';
type CypherSearchMessageProps = {
    // message?: string;
    // showMessage: boolean;
    messageState: {
        showMessage: boolean;
        message?: string;
    };
    clearMessage: () => void;
};

const CypherSearchMessage = (props: CypherSearchMessageProps) => {
    const { clearMessage, messageState } = props;
    const { showMessage, message } = messageState;

    const nodeRef = useRef(null);

    const duration = 300;

    const defaultStyle = {
        transition: `opacity ${duration}ms ease-in-out, transform ${duration}ms`,
        opacity: 0,
        transform: 'scale(1)',
    };

    const transitionStyles: any = {
        entering: { opacity: 1, transform: 'translateX(0) scale(0.96)' },
        entered: { opacity: 1 },
        exiting: { opacity: 0, transform: 'scale(0.9)' },
        exited: { opacity: 0 },
    };

    const handleEntered = () => {
        setTimeout(() => {
            clearMessage();
        }, 5000);
    };

    return (
        <div className='w-full'>
            <Transition nodeRef={nodeRef} in={showMessage} timeout={duration} onEntered={handleEntered}>
                {(state) => (
                    <div
                        ref={nodeRef}
                        style={{
                            ...defaultStyle,
                            ...transitionStyles[state],
                        }}
                        className='leading-none'>
                        {message}
                    </div>
                )}
            </Transition>
        </div>
    );
};

export default CypherSearchMessage;
