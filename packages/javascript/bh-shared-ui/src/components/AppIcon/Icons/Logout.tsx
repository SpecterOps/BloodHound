import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const Logout: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG name='logout ' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' {...props}>
            <BasePath d='M17 8L15.59 9.41L17.17 11H9V13H17.17L15.59 14.58L17 16L21 12L17 8ZM5 5H12V3H5C3.9 3 3 3.9 3 5V19C3 20.1 3.9 21 5 21H12V19H5V5Z' />
        </BaseSVG>
    );
};

export default Logout;
