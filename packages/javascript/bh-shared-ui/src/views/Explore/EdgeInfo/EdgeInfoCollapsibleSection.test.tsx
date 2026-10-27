// Copyright 2025 Specter Ops, Inc.
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
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../../../test-utils';
import EdgeInfoCollapsibleSection from './EdgeInfoCollapsibleSection';

describe('EdgeInfoCollapsibleSection', () => {
    it('supports controlled expansion, keyboard collapse, and unmounting closed content', async () => {
        const user = userEvent.setup();
        const onChange = vi.fn();
        const { rerender } = render(
            <EdgeInfoCollapsibleSection label='Relationship Information' isExpanded={false} onChange={onChange}>
                <span>Edge properties</span>
            </EdgeInfoCollapsibleSection>
        );
        const trigger = screen.getByRole('button', { name: 'Relationship Information' });
        expect(trigger).toHaveAttribute('aria-expanded', 'false');
        expect(screen.queryByText('Edge properties')).not.toBeInTheDocument();

        await user.click(trigger);
        expect(onChange).toHaveBeenLastCalledWith(true);
        expect(trigger).toHaveAttribute('aria-expanded', 'false');

        rerender(
            <EdgeInfoCollapsibleSection label='Relationship Information' isExpanded onChange={onChange}>
                <span>Edge properties</span>
            </EdgeInfoCollapsibleSection>
        );
        expect(trigger).toHaveAttribute('aria-expanded', 'true');
        expect(screen.getByText('Edge properties')).toBeVisible();

        trigger.focus();
        await user.keyboard('{Enter}');
        expect(onChange).toHaveBeenLastCalledWith(false);

        rerender(
            <EdgeInfoCollapsibleSection label='Relationship Information' isExpanded={false} onChange={onChange}>
                <span>Edge properties</span>
            </EdgeInfoCollapsibleSection>
        );
        expect(trigger).toHaveAttribute('aria-expanded', 'false');
        await waitFor(() => expect(screen.queryByText('Edge properties')).not.toBeInTheDocument());
    });
});
