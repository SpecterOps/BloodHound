import { cn, getCssColor } from 'components/utils';
import { ColorOptions } from '../../types';
import { useEffect, useState } from 'react';
import { clampNumber, getCircumference } from './utils';

interface Props extends React.HTMLAttributes<HTMLDivElement> {
    value: number;
    color: ColorOptions;
    hideAnimation?: boolean;
}

const radius = 24;
const strokeWidth = 16;
const containerDimension = 32;

// TSX structure and progress-ring__circle class are pulled from https://gist.github.com/eYinka/873be69fae3ef27b103681b8a9f5e379
function RadialGauge(props: Props) {
    const { value: propsValue, color, hideAnimation = false, ...rest } = props;

    const clampedValue = clampNumber(propsValue, 0, 100);

    const [value, setValue] = useState(hideAnimation ? clampedValue : 0);

    useEffect(() => {
        setValue(clampedValue);
    }, [clampedValue]);

    const circumference = getCircumference(radius);

    const fillColor = getCssColor(color);

    return (
        <div className='relative' style={{ width: containerDimension, height: containerDimension }} {...rest}>
            <svg className='w-full h-full' viewBox='0 0 100 100'>
                <circle
                    className='text-neutral-light-4 stroke-current'
                    strokeWidth={strokeWidth}
                    cx='50'
                    cy='50'
                    r={radius}
                    fill='transparent'
                />
                <circle
                    style={{ color: fillColor }}
                    className={cn('transform -rotate-90 origin-center stroke-current', {
                        'transition-[stroke-dashoffset] duration-300': !hideAnimation,
                    })}
                    strokeWidth={strokeWidth}
                    strokeLinecap='round'
                    cx='50'
                    cy='50'
                    r={radius}
                    fill='transparent'
                    strokeDasharray={circumference}
                    // Must use pixel for this calculation to support Firefox
                    strokeDashoffset={`calc(${circumference}px - (${circumference}px * ${value}) / 100)`}
                />
            </svg>
        </div>
    );
}

export { RadialGauge };
