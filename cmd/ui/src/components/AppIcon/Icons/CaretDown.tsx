import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const CaretDown: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='caret-down'
            width='10'
            height='6'
            viewBox='0 0 10 6'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath d='M0 0.977539L5 5.97754L10 0.977539H0Z' />
        </BaseSVG>
    );
};
