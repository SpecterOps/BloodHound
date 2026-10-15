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
import { createContext, type ReactNode, useContext } from 'react';

type StepperOrientation = 'horizontal' | 'vertical';
type StepState = 'active' | 'completed' | 'inactive' | 'loading';
type StepIndicators = Partial<Record<StepState, ReactNode>>;

interface StepperContextValue {
    id: string;
    tabStopStep?: number;
    activeStep: number;
    setActiveStep: (step: number) => void;
    stepsCount: number;
    orientation: StepperOrientation;
    registerStep: (step: number, isDisabled: boolean) => () => void;
    registerTrigger: (node: HTMLButtonElement, step: number, isDisabled: boolean) => () => void;
    focusNext: (currentNode: HTMLButtonElement) => void;
    focusPrev: (currentNode: HTMLButtonElement) => void;
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

/** Returns the current step, registered step count, and navigation controls. */
function useStepper() {
    const ctx = useContext(StepperContext);
    if (!ctx) throw new Error('useStepper must be used within a Stepper');
    return ctx;
}

/** Returns the state of the enclosing StepperItem. */
function useStepItem() {
    const ctx = useContext(StepItemContext);
    if (!ctx) throw new Error('useStepItem must be used within a StepperItem');
    return ctx;
}

export { StepItemContext, StepperContext, useStepItem, useStepper };
export type { StepIndicators, StepperContextValue, StepperOrientation, StepState };
