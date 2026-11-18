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
import { createRef } from 'react';
import { Button } from '../Button';
import { Input } from '../Input';
import { ButtonGroup, ButtonGroupSeparator, ButtonGroupText } from './ButtonGroup';

describe('ButtonGroup', () => {
    it('preserves native tab order, activation, and disabled buttons', async () => {
        const user = userEvent.setup();
        const onSave = vi.fn();
        const onSaveAs = vi.fn();
        render(
            <ButtonGroup aria-label='Save query actions'>
                <Button onClick={onSave}>Save</Button>
                <Button disabled>Unavailable</Button>
                <Button onClick={onSaveAs}>Save As</Button>
            </ButtonGroup>
        );

        expect(screen.getByRole('group', { name: 'Save query actions' })).toBeInTheDocument();
        await user.tab();
        expect(screen.getByRole('button', { name: 'Save' })).toHaveFocus();
        await user.keyboard('{Enter}');
        expect(onSave).toHaveBeenCalledOnce();
        await user.tab();
        expect(screen.getByRole('button', { name: 'Save As' })).toHaveFocus();
        await user.keyboard(' ');
        expect(onSaveAs).toHaveBeenCalledOnce();
    });

    it('composes a rendered label and forwards refs to group parts', () => {
        const groupRef = createRef<HTMLDivElement>();
        const textRef = createRef<HTMLDivElement>();
        const separatorRef = createRef<HTMLDivElement>();
        render(
            <ButtonGroup ref={groupRef} aria-label='Node search' orientation='vertical'>
                <ButtonGroupText ref={textRef} render={<label htmlFor='nodes'>Nodes</label>}>
                    Nodes
                </ButtonGroupText>
                <Input id='nodes' />
                <ButtonGroupSeparator ref={separatorRef} orientation='horizontal' />
                <Button>Search</Button>
            </ButtonGroup>
        );

        expect(groupRef.current).toBe(screen.getByRole('group', { name: 'Node search' }));
        expect(groupRef.current).toHaveAttribute('data-orientation', 'vertical');
        expect(textRef.current?.tagName).toBe('LABEL');
        expect(screen.getByRole('textbox', { name: 'Nodes' })).toBeInTheDocument();
        expect(separatorRef.current).toBe(screen.getByRole('separator'));
        expect(separatorRef.current).toHaveAttribute('aria-orientation', 'horizontal');
    });
});
