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
import * as React from 'react';
import { Input } from '../Input';
import { cn, cnWithState } from '../utils';
import {
    sliderControlStyles,
    sliderIndicatorStyles,
    sliderRootStyles,
    sliderThumbDotStyles,
    sliderThumbStyles,
    sliderTrackStyles,
} from './Slider.styles';

type SliderRootProps = Omit<BaseSliderRootProps<number>, 'children'>;
type SliderChangeEventDetails = Parameters<NonNullable<SliderRootProps['onValueChange']>>[1];

type SliderProps = Omit<SliderRootProps, 'orientation' | 'minStepsBetweenValues' | 'thumbCollisionBehavior'> & {
    /** Accessible label applied to the slider thumb. */
    thumbAriaLabel?: string;
    /** Values at which to render labeled dot ticks. Values outside the range are ignored. */
    ticks?: readonly number[];
    /** Show an editable value field alongside the slider. Commits on blur or Enter. */
    showValue?: boolean;
    /** Suffix displayed after tick labels and the value field. */
    valueSuffix?: string;
    /** Accessible label for the optional value field. Defaults to the thumb label. */
    valueAriaLabel?: string;
};

/**
 * A slider for selecting a value, built on Base UI.
 */
const Slider = React.forwardRef<HTMLDivElement, SliderProps>(
    (
        {
            className,
            thumbAriaLabel,
            ticks = [],
            showValue = false,
            valueSuffix = '%',
            valueAriaLabel,
            value,
            defaultValue,
            onValueChange,
            onValueCommitted,
            disabled,
            min = 0,
            max = 100,
            step = 1,
            'aria-describedby': ariaDescribedBy,
            ...props
        },
        ref
    ) => {
        const [uncontrolledValue, setUncontrolledValue] = React.useState(defaultValue ?? min);
        const [draftValue, setDraftValue] = React.useState<string | null>(null);
        const errorId = React.useId();
        const suffixId = React.useId();

        const validatedStep = Number.isFinite(step) && step > 0 ? step : 1;
        const displayValue = Math.min(max, Math.max(min, value ?? uncontrolledValue));
        const tickValues = [...new Set(ticks)].filter((tick) => Number.isFinite(tick) && tick >= min && tick <= max);
        const invalidValue =
            draftValue !== null &&
            (draftValue.trim() === '' ||
                !Number.isFinite(Number(draftValue)) ||
                Number(draftValue) < min ||
                Number(draftValue) > max);
        const describedBy =
            [ariaDescribedBy, invalidValue ? errorId : undefined].filter(Boolean).join(' ') || undefined;

        const handleValueChange = (nextValue: number, eventDetails: SliderChangeEventDetails) => {
            setDraftValue(null);
            onValueChange?.(nextValue, eventDetails);

            if (value === undefined && !eventDetails.isCanceled) {
                setUncontrolledValue(nextValue);
            }
        };

        const commitValue = (event: React.FocusEvent<HTMLInputElement> | React.KeyboardEvent<HTMLInputElement>) => {
            if (draftValue === null || invalidValue || disabled || event.currentTarget.disabled) return;

            const nextValue = Math.min(
                max,
                Math.max(
                    min,
                    Number(
                        (min + Math.round((Number(draftValue) - min) / validatedStep) * validatedStep).toPrecision(12)
                    )
                )
            );
            let isCanceled = false;
            let isPropagationAllowed = false;
            const eventDetails: SliderChangeEventDetails = {
                reason: 'input-change',
                event: event.nativeEvent,
                activeThumbIndex: 0,
                trigger: event.currentTarget,
                get isCanceled() {
                    return isCanceled;
                },
                get isPropagationAllowed() {
                    return isPropagationAllowed;
                },
                cancel: () => {
                    isCanceled = true;
                },
                allowPropagation: () => {
                    isPropagationAllowed = true;
                },
            };

            if (nextValue !== displayValue) {
                handleValueChange(nextValue, eventDetails);
                if (!eventDetails.isCanceled) {
                    onValueCommitted?.(nextValue, { reason: 'input-change', event: event.nativeEvent });
                }
            }
            setDraftValue(null);
        };

        const isActive = displayValue > min;

        return (
            <SliderPrimitive.Root
                ref={ref}
                className={cnWithState(sliderRootStyles, showValue && invalidValue && 'mb-5', className)}
                min={min}
                max={max}
                step={validatedStep}
                disabled={disabled}
                value={displayValue}
                onValueChange={handleValueChange}
                onValueCommitted={onValueCommitted}
                {...props}>
                <div className='min-w-0 flex-1'>
                    <SliderPrimitive.Control className={sliderControlStyles}>
                        <SliderPrimitive.Track className={sliderTrackStyles}>
                            <SliderPrimitive.Indicator className={sliderIndicatorStyles({ active: isActive })} />
                        </SliderPrimitive.Track>
                        <SliderPrimitive.Thumb
                            aria-label={thumbAriaLabel}
                            aria-describedby={describedBy}
                            className={sliderThumbStyles({ active: isActive })}>
                            <span aria-hidden data-slider-dot className={sliderThumbDotStyles({ active: isActive })} />
                        </SliderPrimitive.Thumb>
                    </SliderPrimitive.Control>
                    {tickValues.length > 0 && (
                        <div
                            aria-hidden='true'
                            className={cn(
                                'relative h-7 text-[10px] leading-4',
                                disabled ? 'text-disabled' : 'text-main'
                            )}>
                            {tickValues.map((tick) => (
                                <span
                                    key={tick}
                                    className='absolute flex -translate-x-1/2 rtl:translate-x-1/2 flex-col items-center gap-1'
                                    style={{ insetInlineStart: `${((tick - min) / (max - min)) * 100}%` }}>
                                    <span
                                        className={cn(
                                            'size-1 rounded-full',
                                            disabled
                                                ? 'bg-disabled'
                                                : tick <= displayValue && isActive
                                                  ? 'bg-[color:var(--slider-current-color,var(--primary))]'
                                                  : 'bg-[color:var(--slider-tick-fill)]'
                                        )}
                                    />
                                    <span>
                                        {tick}
                                        {valueSuffix}
                                    </span>
                                </span>
                            ))}
                        </div>
                    )}
                </div>
                {showValue && (
                    <SliderPrimitive.Value
                        className='relative shrink-0'
                        htmlFor={undefined}
                        render={(valueProps, { disabled: isDisabled }) => (
                            <div {...valueProps}>
                                <Input
                                    type='text'
                                    inputMode='decimal'
                                    variant='outlined'
                                    aria-label={valueAriaLabel ?? `${thumbAriaLabel ?? 'Slider'} value`}
                                    aria-describedby={
                                        [describedBy, valueSuffix ? suffixId : undefined].filter(Boolean).join(' ') ||
                                        undefined
                                    }
                                    aria-invalid={invalidValue || undefined}
                                    disabled={isDisabled}
                                    value={draftValue ?? String(displayValue)}
                                    onChange={(event) => setDraftValue(event.target.value)}
                                    onBlur={commitValue}
                                    onKeyDown={(event) => {
                                        if (event.key === 'Enter') {
                                            if (event.nativeEvent.isComposing || event.nativeEvent.keyCode === 229)
                                                return;
                                            event.preventDefault();
                                            commitValue(event);
                                        } else if (event.key === 'Escape') {
                                            setDraftValue(null);
                                        }
                                    }}
                                    className={cn('h-7 w-16 px-2 py-1 text-xs', valueSuffix && 'pr-5')}
                                />
                                {valueSuffix && (
                                    <span
                                        id={suffixId}
                                        className={cn(
                                            'pointer-events-none absolute end-2 top-1.5 text-xs',
                                            isDisabled ? 'text-disabled' : 'text-main'
                                        )}>
                                        {valueSuffix}
                                    </span>
                                )}
                            </div>
                        )}
                    />
                )}
                {showValue && invalidValue && (
                    <span
                        id={errorId}
                        role='alert'
                        className='absolute start-0 top-full text-xs text-status-error-main'>
                        Enter a value between {min} and {max}.
                    </span>
                )}
            </SliderPrimitive.Root>
        );
    }
);

Slider.displayName = 'Slider';

export { Slider, type SliderProps };
