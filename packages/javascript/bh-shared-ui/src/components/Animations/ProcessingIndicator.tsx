import { FC } from 'react';

const ProcessingIndicator: FC<{ title: string }> = ({ title }) => {
    return (
        <div className='inline-flex items-center'>
            <span className='animate-pulse'>{title}</span>
            <span className='animate-pulse'>.</span>
            <span className='animate-pulse' style={{ animationDelay: '0.2s' }}>
                .
            </span>
            <span className='animate-pulse' style={{ animationDelay: '0.4s' }}>
                .
            </span>
        </div>
    );
};

export default ProcessingIndicator;
