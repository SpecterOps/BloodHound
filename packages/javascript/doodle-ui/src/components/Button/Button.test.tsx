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

import '@testing-library/jest-dom';
import matchers from '@testing-library/jest-dom/matchers';
import { render, screen } from '@testing-library/react';
import { expect } from 'vitest';
import { Button, ButtonVariants, IconButton, TextButton } from './Button';

expect.extend(matchers);

describe('Button', () => {
    it.each([
        ['primary', 'border border-transparent bg-primary text-text-contrast'],
        ['secondary', 'border-border bg-elevation-1 text-text-main'],
        ['tertiary', 'border-brand-orange bg-elevation-1 text-text-main'],
    ] as const)('applies the %s visual treatment', (variant, expectedClasses) => {
        render(<Button variant={variant}>{variant}</Button>);

        expect(screen.getByRole('button', { name: variant })).toHaveClass(
            'rounded',
            'p-2',
            'text-sm/5',
            'hover:bg-secondary',
            'hover:text-text-contrast',
            'focus-visible:bg-secondary',
            'focus-visible:text-text-contrast',
            'active:bg-primary-variant',
            'active:text-text-contrast',
            ...expectedClasses.split(' ')
        );
        expect(screen.getByRole('button', { name: variant })).not.toHaveClass('shadow-outer-1');
    });

    it.each(['small', 'medium', 'large'] as const)('keeps uniform geometry for the deprecated %s size', (size) => {
        expect(ButtonVariants({ size })).toContain('rounded p-2 text-sm/5');
    });

    it('keeps deprecated variants free of contained-button geometry', () => {
        expect(ButtonVariants({ variant: 'transparent' })).toContain('rounded-3xl');
        expect(ButtonVariants({ variant: 'transparent' })).not.toContain('rounded p-2');
        expect(ButtonVariants({ variant: 'icon' })).not.toContain('rounded p-2');
    });

    it('defaults to a non-submitting button', () => {
        render(<Button>Save</Button>);

        expect(screen.getByRole('button', { name: 'Save' })).toHaveAttribute('type', 'button');
    });

    it('keeps TextButton and IconButton geometry independent', () => {
        render(
            <>
                <TextButton>Help</TextButton>
                <IconButton aria-label='Settings' />
            </>
        );

        expect(screen.getByRole('button', { name: 'Help' })).toHaveClass('rounded-3xl');
        expect(screen.getByRole('button', { name: 'Settings' })).toHaveClass('rounded-full');
    });
});
