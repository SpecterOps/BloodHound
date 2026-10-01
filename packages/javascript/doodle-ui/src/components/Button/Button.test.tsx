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

import { render, screen } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
    it('defaults to type button', () => {
        render(<Button>Button</Button>);

        expect(screen.getByRole('button', { name: 'Button' }).getAttribute('type')).toBe('button');
    });

    it('accepts type submit', () => {
        render(<Button type='submit'>Submit</Button>);

        expect(screen.getByRole('button', { name: 'Submit' }).getAttribute('type')).toBe('submit');
    });

    it('preserves the primary variant colors and active state', () => {
        render(<Button>Primary</Button>);

        const button = screen.getByRole('button', { name: 'Primary' });

        expect(button.classList.contains('text-common-white')).toBe(true);
        expect(button.classList.contains('dark:text-common-dark')).toBe(true);
        expect(button.classList.contains('active:bg-[#0D0A30]')).toBe(true);
        expect(button.classList.contains('dark:active:bg-[#8D8BF8]')).toBe(true);
    });

    it('preserves the secondary variant colors and active state without default opacity', () => {
        render(<Button variant='secondary'>Secondary</Button>);

        const button = screen.getByRole('button', { name: 'Secondary' });

        expect(button.classList.contains('text-common-dark')).toBe(true);
        expect(button.classList.contains('dark:text-common-white')).toBe(true);
        expect(button.classList.contains('active:bg-secondary-btn-active-fill')).toBe(true);
        expect(button.classList.contains('active:text-common-dark')).toBe(true);
        expect(button.classList.contains('dark:active:text-common-white')).toBe(true);
        expect(button.classList.contains('disabled:opacity-50')).toBe(true);
        expect(button.classList.contains('opacity-50')).toBe(false);
    });
});
