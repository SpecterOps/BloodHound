// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
import { useEffect, useState } from 'react';
import { ColorOptions } from '../../types';
import { cn, getCssColor } from '../utils';
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
