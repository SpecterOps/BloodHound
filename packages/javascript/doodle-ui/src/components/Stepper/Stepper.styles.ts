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

import { cn } from '../utils';

export const basicStepperClasses = 'flex min-h-0 flex-col gap-8';

export const basicStepperTextClasses =
    'flex flex-col max-sm:group-data-[orientation=horizontal]/stepper-nav:text-center';

export const basicStepperPanelClasses = 'min-h-0 flex-1 overflow-y-auto overscroll-contain';

export const stepperClasses = 'w-full font-sans text-main';

export const stepperItemClasses =
    'group/step flex flex-auto items-center last:flex-none group-data-[orientation=horizontal]/stepper-nav:flex-row group-data-[orientation=vertical]/stepper-nav:flex-col';

export const stepperIndicatorClasses =
    'group/indicator flex size-10 shrink-0 items-center justify-center rounded border border-transparent text-base font-bold leading-6 data-[state=active]:border-secondary';

export const stepperIndicatorContentClasses = cn(
    'flex size-8 items-center justify-center rounded border border-transparent group-data-[state=inactive]/indicator:border-[color:var(--stepper-inactive-border)]',
    'group-data-[state=completed]/indicator:border-primary group-data-[state=completed]/indicator:bg-primary group-data-[state=completed]/indicator:text-inverse',
    'group-data-[state=active]/indicator:rounded-sm group-data-[state=active]/indicator:border-secondary group-data-[state=active]/indicator:bg-secondary group-data-[state=active]/indicator:text-inverse'
);

export const stepperSeparatorClasses = cn(
    'max-sm:group-data-[orientation=horizontal]/stepper-nav:mt-6 max-sm:group-data-[orientation=horizontal]/stepper-nav:self-start',
    'bg-neutral-300',
    'group-data-[orientation=horizontal]/stepper-nav:h-px group-data-[orientation=horizontal]/stepper-nav:min-w-4 group-data-[orientation=horizontal]/stepper-nav:flex-1',
    'group-data-[orientation=vertical]/stepper-nav:h-12 group-data-[orientation=vertical]/stepper-nav:w-px'
);

export const stepperTitleClasses = 'whitespace-nowrap text-base font-normal leading-4';

export const stepperNavClasses = cn(
    'group/stepper-nav flex items-center overflow-x-auto',
    'data-[orientation=horizontal]:w-full data-[orientation=horizontal]:flex-row',
    'data-[orientation=vertical]:flex-col',
    'max-sm:data-[orientation=horizontal]:items-start'
);

export const stepperTriggerClasses = cn(
    'inline-flex shrink-0 cursor-pointer items-center gap-1.5 rounded px-2 py-1 text-left outline-none focus-visible:z-10 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-secondary disabled:pointer-events-none disabled:opacity-60',
    'max-sm:group-data-[orientation=horizontal]/stepper-nav:flex-col max-sm:group-data-[orientation=horizontal]/stepper-nav:text-center'
);
