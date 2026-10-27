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

import { cva } from 'class-variance-authority';
import { cn } from '../utils';

export const DialogOverlayVariants = cva(
    'fixed inset-0 bg-[color:var(--dialog-overlay-fill)] data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
    { variants: { blurBackground: { true: 'backdrop-blur-sm' }, allowNav: { true: 'z-50', false: 'z-[1410]' } } }
);

export const dialogTriggerClasses = 'focus:outline-none focus-visible:focus-ring';

export const dialogContentClasses = cn(
    'flex flex-col overflow-hidden fixed gap-4 p-6 w-full shadow-lg max-w-lg max-h-[calc(100vh-32px)]',
    'left-[50%] top-[50%] translate-x-[-50%] translate-y-[-50%]',
    'duration-200',
    'data-[state=open]:animate-in',
    'data-[state=closed]:animate-out data-[state=closed]:fade-out-0',
    'data-[state=open]:fade-in-0',
    'data-[state=closed]:zoom-out-95',
    'data-[state=open]:zoom-in-95',
    'data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%]',
    'data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]',
    'rounded-md bg-[color:var(--dialog-fill)] text-[color:var(--dialog-text)] z-[1500]',
    'focus:outline-none focus-visible:focus-ring'
);

export const dialogBodyClasses =
    'flex flex-col gap-4 flex-auto min-h-0 overflow-y-auto px-[calc(var(--focus-ring-width)+var(--focus-ring-offset-width))] -mx-[calc(var(--focus-ring-width)+var(--focus-ring-offset-width))]';

export const dialogActionsClasses = 'flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2';

// Match MUI maxWidth values.
export const dialogMaxWidthClasses = {
    xl: 'max-w-[1536px]',
    lg: 'max-w-[1200px]',
    md: 'max-w-[900px]',
    sm: 'max-w-[600px]',
    xs: 'max-w-[444px]',
};
