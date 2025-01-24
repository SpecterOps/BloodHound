import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const Equals: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='equals'
            xmlns='http://www.w3.org/2000/svg'
            version='1.1'
            x='0px'
            y='0px'
            viewBox='0 0 100 100'
            enableBackground='new 0 0 100 100'
            xmlSpace='preserve'
            {...props}>
            <BasePath d='M91.334,33.803c0,3.276-2.656,5.932-5.933,5.932H13.517c-3.276,0-5.933-2.656-5.933-5.932l0,0  c0-3.276,2.656-5.933,5.933-5.933h71.885C88.678,27.871,91.334,30.527,91.334,33.803L91.334,33.803z' />
            <BasePath d='M91.334,66.605c0,3.275-2.656,5.932-5.933,5.932H13.517c-3.276,0-5.933-2.656-5.933-5.932l0,0  c0-3.276,2.656-5.933,5.933-5.933h71.885C88.678,60.673,91.334,63.329,91.334,66.605L91.334,66.605z' />
        </BaseSVG>
    );
};
