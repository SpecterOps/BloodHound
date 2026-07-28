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
import matchers from '@testing-library/jest-dom/matchers';
import { screen } from '@testing-library/react';
import { expect } from 'vitest';
import { render } from '../../utils';
import { CardDescription, CardTitle } from './Card';

expect.extend(matchers);

describe('Card typography', () => {
    it('uses the semantic h3 typography foundation for titles', () => {
        render(<CardTitle>Title</CardTitle>);

        expect(screen.getByRole('heading', { level: 3 })).toHaveClass(
            'font-heading',
            'text-xl',
            'font-bold',
            'leading-[1.375rem]'
        );
    });

    it('uses semantic muted body2 typography for descriptions', () => {
        render(<CardDescription>Description</CardDescription>);

        expect(screen.getByText('Description')).toHaveClass(
            'font-sans',
            'text-sm',
            'font-normal',
            'leading-[1.375rem]',
            'text-text-muted',
            'dark:text-text-main'
        );
    });
});
