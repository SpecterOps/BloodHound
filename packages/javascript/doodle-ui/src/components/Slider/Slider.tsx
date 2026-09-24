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
import { Slider as SliderPrimitive, type SliderRootProps as BaseSliderRootProps } from '@base-ui/react/slider';
import { cva } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';

const sliderRootStyles =
    'relative flex w-full touch-none select-none items-center data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50';

const sliderControlStyles = 'group/slider relative flex w-full grow items-center py-3';

const sliderTrackStyles =
    'relative h-1 w-full grow overflow-hidden rounded-full bg-neutral-200 data-[disabled]:bg-disabled';

const sliderIndicatorStyles = cva(
    'h-full rounded-full group-has-[:focus-visible]/slider:bg-secondary data-[disabled]:bg-disabled',
    {
        variants: {
            active: {
                true: 'bg-primary',
                false: 'bg-neutral-200',
            },
        },
        defaultVariants: {
            active: false,
        },
    }
);

const sliderThumbStyles = cva(
    'group/thumb relative size-4 rounded-full border border-neutral-300 bg-neutral-100 shadow-outer-1 dark:has-[:focus-visible]:bg-neutral-400 data-[disabled]:bg-disabled',
    {
        variants: {
            active: {
                true: 'dark:bg-common-white',
                false: '',
            },
        },
        defaultVariants: {
            active: false,
        },
    }
);

const sliderThumbDotStyles = cva(
    'pointer-events-none absolute left-1/2 top-1/2 size-1.5 -translate-x-1/2 -translate-y-1/2 rounded-full group-has-[:focus-visible]/thumb:block group-has-[:focus-visible]/thumb:bg-secondary group-data-[disabled]/thumb:hidden',
    {
        variants: {
            active: {
                true: 'block bg-primary',
                false: 'hidden',
            },
        },
        defaultVariants: {
            active: false,
        },
    }
);

type SliderRootProps = Omit<BaseSliderRootProps<number>, 'children'>;
type SliderChangeEventDetails = Parameters<NonNullable<SliderRootProps['onValueChange']>>[1];

type SliderProps = Omit<SliderRootProps, 'orientation' | 'minStepsBetweenValues' | 'thumbCollisionBehavior'> & {
    /** Accessible label applied to the slider thumb. */
    thumbAriaLabel?: string;
};

/**
 * A slider for selecting a value, built on Base UI.
 */
const Slider = React.forwardRef<HTMLDivElement, SliderProps>(
    (
        {
            className,
            thumbAriaLabel,
            value,
            defaultValue,
            onValueChange,
            min = 0,
            'aria-describedby': ariaDescribedBy,
            ...props
        },
        ref
    ) => {
        const [uncontrolledValue, setUncontrolledValue] = React.useState(defaultValue);

        const displayValue = value ?? uncontrolledValue;

        const handleValueChange = (nextValue: number, eventDetails: SliderChangeEventDetails) => {
            onValueChange?.(nextValue, eventDetails);

            if (value === undefined && !eventDetails.isCanceled) {
                setUncontrolledValue(nextValue);
            }
        };

        const isActive = (displayValue ?? min) > min;

        return (
            <SliderPrimitive.Root
                ref={ref}
                className={(state) =>
                    cn(sliderRootStyles, typeof className === 'function' ? className(state) : className)
                }
                min={min}
                value={value}
                defaultValue={defaultValue}
                onValueChange={handleValueChange}
                {...props}>
                <SliderPrimitive.Control className={sliderControlStyles}>
                    <SliderPrimitive.Track className={sliderTrackStyles}>
                        <SliderPrimitive.Indicator className={sliderIndicatorStyles({ active: isActive })} />
                    </SliderPrimitive.Track>
                    <SliderPrimitive.Thumb
                        aria-label={thumbAriaLabel}
                        aria-describedby={ariaDescribedBy}
                        className={sliderThumbStyles({ active: isActive })}>
                        <span aria-hidden className={sliderThumbDotStyles({ active: isActive })} />
                    </SliderPrimitive.Thumb>
                </SliderPrimitive.Control>
            </SliderPrimitive.Root>
        );
    }
);

Slider.displayName = 'Slider';

export { Slider, type SliderProps };
