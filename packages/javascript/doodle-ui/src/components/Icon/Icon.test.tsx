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
import userEvent from '@testing-library/user-event';
import { AppIcon } from '../../styleguide/components/AppIcons/AppIcons';
import { Icon } from './Icon';

describe('Icon', () => {
    it('renders an AppIcon without button semantics', () => {
        const { container } = render(
            <Icon aria-label='Information'>
                <AppIcon.Info />
            </Icon>
        );

        expect(container.querySelector('svg')?.getAttribute('aria-label')).toBe('Information');
        expect(screen.queryByRole('button')).toBeNull();
    });

    it('uses its accessible label as its tooltip', async () => {
        const user = userEvent.setup();
        const { container } = render(
            <Icon aria-label='Filter options'>
                <AppIcon.Info />
            </Icon>
        );

        await user.hover(container.querySelector('svg') as SVGSVGElement);

        expect((await screen.findByRole('tooltip')).textContent).toBe('Filter options');
    });

    it('does not render a tooltip when it is disabled', async () => {
        const user = userEvent.setup();
        const { container } = render(
            <Icon aria-label='Information' hideTooltip>
                <AppIcon.Info />
            </Icon>
        );

        await user.hover(container.querySelector('svg') as SVGSVGElement);

        expect(screen.queryByRole('tooltip')).toBeNull();
    });
});
