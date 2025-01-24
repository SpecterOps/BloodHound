import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const BarChart: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG name='bar-chart' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' {...props}>
            <BasePath d='M19 3H5C3.9 3 3 3.9 3 5V19C3 20.1 3.9 21 5 21H19C20.1 21 21 20.1 21 19V5C21 3.9 20.1 3 19 3ZM19 19H5V5H19V19ZM7 10H9V17H7V10ZM11 7H13V17H11V7ZM15 13H17V17H15V13Z' />
        </BaseSVG>
    );
};

export default BarChart;
