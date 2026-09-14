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
});
