import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const EclipseCircle: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='eclipse-circle'
            version='1.1'
            xmlns='http://www.w3.org/2000/svg'
            viewBox='0 0 768 768'
            {...props}>
            <g id='eclipse-circle'></g>
            <BasePath d='M636 384c0-139.126-112.875-252-252-252v504c139.125 0 252-112.875 252-252zM47.999 384c0-89.113 35.399-174.576 98.413-237.588s148.475-98.413 237.588-98.413c89.114 0 174.576 35.399 237.589 98.413 63.009 63.012 98.412 148.475 98.412 237.588 0 89.114-35.4 174.576-98.412 237.589-63.013 63.009-148.475 98.412-237.589 98.412s-174.576-35.4-237.588-98.412c-63.011-63.013-98.413-148.475-98.413-237.589z' />
        </BaseSVG>
    );
};

export default EclipseCircle;
