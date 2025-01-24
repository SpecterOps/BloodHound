import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const MagnifyingGlass: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG viewBox='0 0 18 18' fill='none' xmlns='http://www.w3.org/2000/svg' name='magnifying-glass' {...props}>
            <BasePath d='M12.755 11.255H11.965L11.685 10.985C12.665 9.84501 13.255 8.36501 13.255 6.75501C13.255 3.16501 10.345 0.255005 6.755 0.255005C3.165 0.255005 0.255005 3.16501 0.255005 6.75501C0.255005 10.345 3.165 13.255 6.755 13.255C8.365 13.255 9.845 12.665 10.985 11.685L11.255 11.965V12.755L16.255 17.745L17.745 16.255L12.755 11.255ZM6.755 11.255C4.26501 11.255 2.255 9.24501 2.255 6.75501C2.255 4.26501 4.26501 2.25501 6.755 2.25501C9.245 2.25501 11.255 4.26501 11.255 6.75501C11.255 9.24501 9.245 11.255 6.755 11.255Z' />
        </BaseSVG>
    );
};

export default MagnifyingGlass;
