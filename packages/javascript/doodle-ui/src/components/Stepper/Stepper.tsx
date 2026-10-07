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
import { mergeProps, useRender } from '@base-ui/react';
import { LoaderCircle } from 'lucide-react';
import {
    type ComponentProps,
    type HTMLAttributes,
    KeyboardEvent as ReactKeyboardEvent,
    type ReactNode,
    useCallback,
    useId,
    useLayoutEffect,
    useRef,
    useState,
} from 'react';
import { Typography } from '../Typography';
import { cn } from '../utils.ts';
import {
    type StepIndicators,
    StepItemContext,
    StepperContext,
    type StepperContextValue,
    type StepperOrientation,
    type StepState,
    useStepItem,
    useStepper,
} from './StepperContext';

interface Step {
    title: string;
    description?: ReactNode;
    content?: ReactNode;
    isLoading?: boolean;
    /** Makes the step unavailable to clicks and navigation requests. Enable it before navigating to it. */
    isDisabled?: boolean;
    /** Allows marking future steps as completed, even if they were never active. */
    isCompleted?: boolean;
}

interface BasicStepperProps extends Omit<StepperProps, 'children'> {
    steps: Step[];
}

/** A preset layout for simple, ordered flows. Use Stepper composition for custom layouts. */
const BasicStepper = ({
    steps,
    className,
    indicators = { loading: <LoaderCircle className='size-4 animate-spin' aria-label='Loading' /> },
    ...props
}: BasicStepperProps) => {
    return (
        <Stepper {...props} indicators={indicators} className={cn('space-y-8', className)}>
            <StepperNav>
                {steps.map((step, index) => (
                    <StepperItem
                        key={step.title}
                        step={index + 1}
                        isLoading={step.isLoading}
                        isDisabled={step.isDisabled}
                        isCompleted={step.isCompleted}>
                        <StepperTrigger>
                            <StepperIndicator>{index + 1}</StepperIndicator>
                            {step.description != null ? (
                                <div className='flex flex-col max-sm:group-data-[orientation=horizontal]/stepper-nav:text-center'>
                                    <StepperTitle>{step.title}</StepperTitle>
                                    <StepperDescription>{step.description}</StepperDescription>
                                </div>
                            ) : (
                                <StepperTitle>{step.title}</StepperTitle>
                            )}
                        </StepperTrigger>

                        {index < steps.length - 1 && <StepperSeparator />}
                    </StepperItem>
                ))}
            </StepperNav>

            <StepperPanel>
                {steps.map((step, index) => (
                    <StepperContent key={index} value={index + 1} className='flex items-center justify-center'>
                        {step.content}
                    </StepperContent>
                ))}
            </StepperPanel>
        </Stepper>
    );
};

interface StepperProps extends HTMLAttributes<HTMLDivElement> {
    /** Initial active step (1-based) for an uncontrolled stepper. */
    defaultValue?: number;
    /** Active step (1-based). The parent must enable the destination before setting this value. */
    value?: number;
    onValueChange?: (value: number) => void;
    orientation?: StepperOrientation;
    indicators?: StepIndicators;
}

const Stepper = ({
    defaultValue = 1,
    value,
    onValueChange,
    orientation = 'horizontal',
    className,
    children,
    indicators = {},
    ...props
}: StepperProps) => {
    const id = useId();
    const [activeStep, setActiveStep] = useState(defaultValue);
    const [stepsCount, setStepsCount] = useState(0);
    const [triggers, setTriggers] = useState<{ node: HTMLButtonElement; step: number; isDisabled: boolean }[]>([]);
    const disabledSteps = useRef(new Map<number, boolean>());

    const registerStep = useCallback((step: number, isDisabled: boolean) => {
        disabledSteps.current.set(step, isDisabled);
        setStepsCount(disabledSteps.current.size);
        return () => {
            disabledSteps.current.delete(step);
            setStepsCount(disabledSteps.current.size);
        };
    }, []);

    // Register/unregister triggers
    const registerTrigger = useCallback((node: HTMLButtonElement, step: number, isDisabled: boolean) => {
        setTriggers((previous) => [...previous, { node, step, isDisabled }]);
        return () => setTriggers((previous) => previous.filter((trigger) => trigger.node !== node));
    }, []);

    const handleSetActiveStep = useCallback(
        (step: number) => {
            if (disabledSteps.current.get(step)) return;
            if (value === undefined) setActiveStep(step);
            onValueChange?.(step);
        },
        [value, onValueChange]
    );

    const currentStep = value ?? activeStep;

    // Resolve DOM order when navigating so keyed insertions and reorders are reflected.
    const getOrderedTriggers = () =>
        triggers
            .filter(({ node, isDisabled }) => node.isConnected && !node.disabled && !isDisabled)
            .sort((first, second) =>
                first.node.compareDocumentPosition(second.node) & Node.DOCUMENT_POSITION_FOLLOWING ? -1 : 1
            );
    const focusAdjacent = (currentNode: HTMLButtonElement, direction: number) => {
        const orderedTriggers = getOrderedTriggers();
        const currentIndex = orderedTriggers.findIndex(({ node }) => node === currentNode);
        if (currentIndex === -1) return;
        const nextIndex = (currentIndex + direction + orderedTriggers.length) % orderedTriggers.length;
        orderedTriggers[nextIndex]?.node.focus();
    };
    const focusNext = (currentNode: HTMLButtonElement) => focusAdjacent(currentNode, 1);
    const focusPrev = (currentNode: HTMLButtonElement) => focusAdjacent(currentNode, -1);
    const focusFirst = () => getOrderedTriggers()[0]?.node.focus();
    const focusLast = () => getOrderedTriggers().at(-1)?.node.focus();
    const enabledTriggers = getOrderedTriggers();
    const tabStopStep = (enabledTriggers.find(({ step }) => step === currentStep) ?? enabledTriggers[0])?.step;

    const contextValue: StepperContextValue = {
        id,
        tabStopStep,
        activeStep: currentStep,
        setActiveStep: handleSetActiveStep,
        stepsCount,
        orientation,
        registerStep,
        registerTrigger,
        focusNext,
        focusPrev,
        focusFirst,
        focusLast,
        indicators,
    };

    return (
        <StepperContext.Provider value={contextValue}>
            <div
                data-slot='stepper'
                className={cn('w-full font-sans text-text-main', className)}
                data-orientation={orientation}
                {...props}>
                {children}
            </div>
        </StepperContext.Provider>
    );
};

interface StepperItemProps extends HTMLAttributes<HTMLDivElement> {
    step: number;
    isCompleted?: boolean;
    isDisabled?: boolean;
    isLoading?: boolean;
}

function StepperItem({
    step,
    isCompleted = false,
    isDisabled = false,
    isLoading = false,
    className,
    children,
    ...props
}: StepperItemProps) {
    const { activeStep, registerStep } = useStepper();

    useLayoutEffect(() => registerStep(step, isDisabled), [registerStep, step, isDisabled]);

    const state: StepState =
        isCompleted || step < activeStep ? 'completed' : activeStep === step ? 'active' : 'inactive';
    const isItemLoading = isLoading && step === activeStep;

    return (
        <StepItemContext.Provider value={{ step, state, isDisabled, isLoading: isItemLoading, isCompleted }}>
            <div
                data-slot='stepper-item'
                className={cn(
                    'group/step flex flex-auto items-center last:flex-none group-data-[orientation=horizontal]/stepper-nav:flex-row group-data-[orientation=vertical]/stepper-nav:flex-col',
                    className
                )}
                data-state={state}
                {...(isItemLoading ? { 'data-loading': true } : {})}
                {...props}>
                {children}
            </div>
        </StepItemContext.Provider>
    );
}

type StepperTriggerProps = useRender.ComponentProps<'button'>;

function StepperTrigger({ className, children, tabIndex, render, disabled = false, ...props }: StepperTriggerProps) {
    const { state, isLoading, step, isDisabled } = useStepItem();
    const {
        id: stepperId,
        tabStopStep,
        setActiveStep,
        activeStep,
        registerTrigger,
        focusNext,
        focusPrev,
        focusFirst,
        focusLast,
    } = useStepper();
    const isSelected = activeStep === step;
    const isTriggerDisabled = isDisabled || disabled;
    const id = `${stepperId}-tab-${step}`;
    const panelId = `${stepperId}-panel-${step}`;

    // Register this trigger for keyboard navigation
    const btnRef = useRef<HTMLButtonElement>(null);
    useLayoutEffect(() => {
        if (btnRef.current) return registerTrigger(btnRef.current, step, isTriggerDisabled);
    }, [registerTrigger, step, isTriggerDisabled]);

    const handleKeyDown = (e: ReactKeyboardEvent<HTMLButtonElement>) => {
        if (isTriggerDisabled) return;
        switch (e.key) {
            case 'ArrowRight':
            case 'ArrowDown':
                e.preventDefault();
                focusNext(e.currentTarget);
                break;
            case 'ArrowLeft':
            case 'ArrowUp':
                e.preventDefault();
                focusPrev(e.currentTarget);
                break;
            case 'Home':
                e.preventDefault();
                focusFirst();
                break;
            case 'End':
                e.preventDefault();
                focusLast();
                break;
            case 'Enter':
            case ' ':
                e.preventDefault();
                setActiveStep(step);
                break;
        }
    };

    const defaultProps = {
        type: 'button' as const,
        role: 'tab',
        id,
        'aria-selected': isSelected,
        'aria-controls': panelId,
        'aria-disabled': isTriggerDisabled || undefined,
        tabIndex: isTriggerDisabled ? -1 : tabIndex ?? (tabStopStep === step ? 0 : -1),
        'data-slot': 'stepper-trigger',
        'data-state': state,
        'data-loading': isLoading,
        className: cn(
            'inline-flex shrink-0 cursor-pointer items-center gap-1.5 rounded px-2 py-1 text-left outline-none focus-visible:z-10 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-secondary disabled:pointer-events-none disabled:opacity-60',
            'max-sm:group-data-[orientation=horizontal]/stepper-nav:flex-col max-sm:group-data-[orientation=horizontal]/stepper-nav:text-center',
            className
        ),
        onClick: () => {
            if (!isTriggerDisabled) setActiveStep(step);
        },
        onKeyDown: handleKeyDown,
        disabled: isTriggerDisabled,
        children,
    };

    return useRender({
        defaultTagName: 'button',
        render,
        ref: btnRef,
        props: mergeProps<'button'>(defaultProps, props),
    });
}

function StepperIndicator({ children, className, ...props }: ComponentProps<'div'>) {
    const { state, isLoading } = useStepItem();
    const { indicators } = useStepper();

    return (
        <div
            {...props}
            data-slot='stepper-indicator'
            data-state={state}
            className={cn(
                'group/indicator flex size-10 shrink-0 items-center justify-center rounded border border-transparent text-base font-bold leading-6 data-[state=active]:border-secondary',
                className
            )}>
            <div
                className={cn(
                    'flex size-8 items-center justify-center rounded border border-transparent group-data-[state=inactive]/indicator:border-[#939597] dark:group-data-[state=inactive]/indicator:border-input-border-default',
                    'group-data-[state=completed]/indicator:border-primary group-data-[state=completed]/indicator:bg-primary group-data-[state=completed]/indicator:text-text-contrast',
                    'group-data-[state=active]/indicator:rounded-sm group-data-[state=active]/indicator:border-secondary group-data-[state=active]/indicator:bg-secondary group-data-[state=active]/indicator:text-text-contrast'
                )}>
                {(isLoading && indicators.loading) || indicators[state] || children}
            </div>
        </div>
    );
}

function StepperSeparator({ className, ...props }: ComponentProps<'div'>) {
    const { state } = useStepItem();

    return (
        <div
            {...props}
            data-slot='stepper-separator'
            data-state={state}
            className={cn(
                'max-sm:group-data-[orientation=horizontal]/stepper-nav:mt-6 max-sm:group-data-[orientation=horizontal]/stepper-nav:self-start',
                'bg-neutral-300 group-data-[orientation=horizontal]/stepper-nav:h-px group-data-[orientation=horizontal]/stepper-nav:min-w-4 group-data-[orientation=horizontal]/stepper-nav:flex-1 group-data-[orientation=vertical]/stepper-nav:h-12 group-data-[orientation=vertical]/stepper-nav:w-px',
                className
            )}
        />
    );
}

function StepperTitle({ children, className, ...props }: ComponentProps<'h3'>) {
    const { state } = useStepItem();

    return (
        <Typography
            {...props}
            variant='h3'
            data-slot='stepper-title'
            data-state={state}
            className={cn('whitespace-nowrap text-base font-normal leading-4', className)}>
            {children}
        </Typography>
    );
}

function StepperDescription({ children, className, ...props }: ComponentProps<'div'>) {
    const { state } = useStepItem();

    return (
        <Typography
            {...props}
            component='div'
            data-slot='stepper-description'
            variant='subtitle2'
            data-state={state}
            className={cn('text-text-muted leading-4', className)}>
            {children}
        </Typography>
    );
}

function StepperNav({ children, className, ...props }: ComponentProps<'div'>) {
    const { activeStep, orientation } = useStepper();

    return (
        <div
            {...props}
            role='tablist'
            aria-orientation={orientation}
            data-slot='stepper-nav'
            data-state={activeStep}
            data-orientation={orientation}
            className={cn(
                'group/stepper-nav flex items-center overflow-x-auto data-[orientation=horizontal]:w-full data-[orientation=horizontal]:flex-row data-[orientation=vertical]:flex-col max-sm:data-[orientation=horizontal]:items-start',
                className
            )}>
            {children}
        </div>
    );
}

function StepperPanel({ children, className, ...props }: ComponentProps<'div'>) {
    const { activeStep } = useStepper();

    return (
        <div {...props} data-slot='stepper-panel' data-state={activeStep} className={cn('w-full', className)}>
            {children}
        </div>
    );
}

interface StepperContentProps extends ComponentProps<'div'> {
    value: number;
    forceMount?: boolean;
}

function StepperContent({ value, forceMount, children, className, ...props }: StepperContentProps) {
    const { activeStep, id: stepperId } = useStepper();
    const isActive = value === activeStep;

    return (
        <div
            // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- Text-only tab panels need a Tab focus target (WAI-ARIA Tabs Pattern).
            tabIndex={0}
            {...props}
            id={`${stepperId}-panel-${value}`}
            role='tabpanel'
            aria-labelledby={`${stepperId}-tab-${value}`}
            data-slot='stepper-content'
            data-state={activeStep}
            className={cn('w-full', className, !isActive && 'hidden')}
            hidden={!isActive}>
            {(isActive || forceMount) && children}
        </div>
    );
}

export {
    BasicStepper,
    Stepper,
    StepperContent,
    StepperDescription,
    StepperIndicator,
    StepperItem,
    StepperNav,
    StepperPanel,
    StepperSeparator,
    StepperTitle,
    StepperTrigger,
    type BasicStepperProps,
    type Step,
    type StepperContentProps,
    type StepperItemProps,
    type StepperProps,
    type StepperTriggerProps,
};
