import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const CaretUp: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='caret-up'
            width='10'
            height='6'
            viewBox='0 0 10 6'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath d='M0 5.97754L5 0.977539L10 5.97754H0Z' />
        </BaseSVG>
    );
};
