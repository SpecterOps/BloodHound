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
    createContext,
    type HTMLAttributes,
    isValidElement,
    Children as ReactChildren,
    type ReactElement,
    KeyboardEvent as ReactKeyboardEvent,
    type ReactNode,
    useCallback,
    useContext,
    useEffect,
    useLayoutEffect,
    useRef,
    useState,
} from 'react';
import { Typography } from '../Typography';
import { cn } from '../utils.ts';

// Types
type StepperOrientation = 'horizontal' | 'vertical';
type StepState = 'active' | 'completed' | 'inactive' | 'loading';
type StepIndicators = Partial<Record<StepState, ReactNode>>;

interface StepperContextValue {
    activeStep: number;
    setActiveStep: (step: number) => void;
    stepsCount: number;
    orientation: StepperOrientation;
    registerStep: (step: number, isDisabled: boolean) => () => void;
    registerTrigger: (node: HTMLButtonElement) => () => void;
    triggerNodes: HTMLButtonElement[];
    focusNext: (currentIdx: number) => void;
    focusPrev: (currentIdx: number) => void;
    focusFirst: () => void;
    focusLast: () => void;
    indicators: StepIndicators;
}

interface StepItemContextValue {
    step: number;
    state: StepState;
    isDisabled: boolean;
    isLoading: boolean;
    isCompleted: boolean;
}

const StepperContext = createContext<StepperContextValue | undefined>(undefined);
const StepItemContext = createContext<StepItemContextValue | undefined>(undefined);

function useStepper() {
    const ctx = useContext(StepperContext);
    if (!ctx) throw new Error('useStepper must be used within a Stepper');
    return ctx;
}

function useStepItem() {
    const ctx = useContext(StepItemContext);
    if (!ctx) throw new Error('useStepItem must be used within a StepperItem');
    return ctx;
}

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
    const [activeStep, setActiveStep] = useState(defaultValue);
    const [triggerNodes, setTriggerNodes] = useState<HTMLButtonElement[]>([]);
    const disabledSteps = useRef(new Map<number, boolean>());

    const registerStep = useCallback((step: number, isDisabled: boolean) => {
        disabledSteps.current.set(step, isDisabled);
        return () => {
            disabledSteps.current.delete(step);
        };
    }, []);

    // Register/unregister triggers
    const registerTrigger = useCallback((node: HTMLButtonElement) => {
        setTriggerNodes((prev) => (prev.includes(node) ? prev : [...prev, node]));
        return () => setTriggerNodes((prev) => prev.filter((trigger) => trigger !== node));
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

    // Keyboard navigation logic
    const focusAdjacent = (currentIdx: number, direction: number) => {
        for (let offset = 1; offset <= triggerNodes.length; offset++) {
            const idx = (currentIdx + direction * offset + triggerNodes.length) % triggerNodes.length;
            const trigger = triggerNodes[idx];
            if (trigger.isConnected && !trigger.disabled) {
                trigger.focus();
                return;
            }
        }
    };
    const focusNext = (currentIdx: number) => focusAdjacent(currentIdx, 1);
    const focusPrev = (currentIdx: number) => focusAdjacent(currentIdx, -1);
    const focusFirst = () => focusAdjacent(triggerNodes.length - 1, 1);
    const focusLast = () => focusAdjacent(0, -1);

    const contextValue: StepperContextValue = {
        activeStep: currentStep,
        setActiveStep: handleSetActiveStep,
        stepsCount: ReactChildren.toArray(children).filter(
            (child): child is ReactElement =>
                isValidElement(child) && (child.type as { displayName?: string }).displayName === 'StepperItem'
        ).length,
        orientation,
        registerStep,
        registerTrigger,
        focusNext,
        focusPrev,
        focusFirst,
        focusLast,
        triggerNodes,
        indicators,
    };

    return (
        <StepperContext.Provider value={contextValue}>
            <div
                role='tablist'
                aria-orientation={orientation}
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
    const { setActiveStep, activeStep, registerTrigger, triggerNodes, focusNext, focusPrev, focusFirst, focusLast } =
        useStepper();
    const isSelected = activeStep === step;
    const isTriggerDisabled = isDisabled || disabled;
    const id = `stepper-tab-${step}`;
    const panelId = `stepper-panel-${step}`;

    // Register this trigger for keyboard navigation
    const btnRef = useRef<HTMLButtonElement>(null);
    useEffect(() => {
        if (btnRef.current) return registerTrigger(btnRef.current);
    }, [registerTrigger]);

    const handleKeyDown = (e: ReactKeyboardEvent<HTMLButtonElement>) => {
        if (isTriggerDisabled) return;
        const myIdx = triggerNodes.indexOf(e.currentTarget);
        switch (e.key) {
            case 'ArrowRight':
            case 'ArrowDown':
                e.preventDefault();
                if (myIdx !== -1 && focusNext) focusNext(myIdx);
                break;
            case 'ArrowLeft':
            case 'ArrowUp':
                e.preventDefault();
                if (myIdx !== -1 && focusPrev) focusPrev(myIdx);
                break;
            case 'Home':
                e.preventDefault();
                if (focusFirst) focusFirst();
                break;
            case 'End':
                e.preventDefault();
                if (focusLast) focusLast();
                break;
            case 'Enter':
            case ' ':
                e.preventDefault();
                setActiveStep(step);
                break;
        }
    };

    const defaultProps = {
        role: 'tab',
        id,
        'aria-selected': isSelected,
        'aria-controls': panelId,
        'aria-disabled': isTriggerDisabled || undefined,
        tabIndex: isTriggerDisabled ? -1 : tabIndex ?? (isSelected ? 0 : -1),
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

function StepperIndicator({ children, className }: ComponentProps<'div'>) {
    const { state, isLoading } = useStepItem();
    const { indicators } = useStepper();

    return (
        <div
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

function StepperSeparator({ className }: ComponentProps<'div'>) {
    const { state } = useStepItem();

    return (
        <div
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

function StepperTitle({ children, className }: ComponentProps<'h3'>) {
    const { state } = useStepItem();

    return (
        <Typography
            variant='h3'
            data-slot='stepper-title'
            data-state={state}
            className={cn('whitespace-nowrap text-base font-normal leading-4', className)}>
            {children}
        </Typography>
    );
}

function StepperDescription({ children, className }: ComponentProps<'div'>) {
    const { state } = useStepItem();

    return (
        <Typography
            component='div'
            data-slot='stepper-description'
            variant='subtitle2'
            data-state={state}
            className={cn('text-text-muted leading-4', className)}>
            {children}
        </Typography>
    );
}

function StepperNav({ children, className }: ComponentProps<'nav'>) {
    const { activeStep, orientation } = useStepper();

    return (
        <nav
            data-slot='stepper-nav'
            data-state={activeStep}
            data-orientation={orientation}
            className={cn(
                'group/stepper-nav flex items-center overflow-x-auto data-[orientation=horizontal]:w-full data-[orientation=horizontal]:flex-row data-[orientation=vertical]:flex-col max-sm:data-[orientation=horizontal]:items-start',
                className
            )}>
            {children}
        </nav>
    );
}

function StepperPanel({ children, className }: ComponentProps<'div'>) {
    const { activeStep } = useStepper();

    return (
        <div data-slot='stepper-panel' data-state={activeStep} className={cn('w-full', className)}>
            {children}
        </div>
    );
}

interface StepperContentProps extends ComponentProps<'div'> {
    value: number;
    forceMount?: boolean;
}

function StepperContent({ value, forceMount, children, className }: StepperContentProps) {
    const { activeStep } = useStepper();
    const isActive = value === activeStep;

    if (!forceMount && !isActive) return null;

    return (
        <div
            data-slot='stepper-content'
            data-state={activeStep}
            className={cn('w-full', className, !isActive && forceMount && 'hidden')}
            hidden={!isActive && forceMount}>
            {children}
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
    useStepItem,
    useStepper,
    type BasicStepperProps,
    type Step,
    type StepperContentProps,
    type StepperItemProps,
    type StepperProps,
    type StepperTriggerProps,
};
