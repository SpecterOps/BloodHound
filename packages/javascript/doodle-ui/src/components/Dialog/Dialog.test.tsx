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
import { Dialog, DialogContent, DialogDescription, DialogTitle } from './Dialog';

expect.extend(matchers);

describe('Dialog typography', () => {
    it('uses the semantic h3 typography foundation for titles', () => {
        render(
            <Dialog open>
                <DialogContent>
                    <DialogTitle>Dialog title</DialogTitle>
                    <DialogDescription>Dialog description</DialogDescription>
                </DialogContent>
            </Dialog>
        );

        expect(screen.getByRole('heading', { name: 'Dialog title' })).toHaveClass(
            'font-heading',
            'text-xl',
            'font-bold',
            'leading-[1.375rem]'
        );
    });
});
