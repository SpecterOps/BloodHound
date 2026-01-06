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
import { setupServer } from 'msw/node';
import { usePZPathParams } from '../../../../hooks';
import { zoneHandlers } from '../../../../mocks';
import { render, screen } from '../../../../test-utils';
import { SelectedDetailsTabs } from './SelectedDetailsTabs';
import SelectedDetailsTabProvider from './SelectedDetailsTabsProvider';

const server = setupServer(...zoneHandlers);

vi.mock('../../../../hooks/usePZParams/');
const mockedUsePathParams = vi.mocked(usePZPathParams);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const SelectedDetailsTabsWrapper = () => {
    return (
        <SelectedDetailsTabProvider>
            <SelectedDetailsTabs />
        </SelectedDetailsTabProvider>
    );
};

describe('Selected Details Tabs', () => {
    describe('Selected Details Tabs - Click Interactions', () => {
        it('switches Rule tab to active/selected when clicked', async () => {
            mockedUsePathParams.mockReturnValue({
                tagId: '1',
                ruleId: '3',
                memberId: undefined,
                tagTypeDisplay: 'Zone',
            } as any);
            render(<SelectedDetailsTabsWrapper />);

            const allTabs = await screen.findAllByRole('tab');
            const zoneTab = allTabs[0];
            const ruleTab = allTabs[1];

            expect(zoneTab).toHaveTextContent(/zone/i);
            expect(zoneTab).toBeEnabled();
            expect(ruleTab).toBeEnabled();

            const user = userEvent.setup();
            await user.click(ruleTab);

            expect(ruleTab).toHaveAttribute('data-state', 'active');
            expect(ruleTab).toHaveAttribute('aria-selected', 'true');
        });
    });
    describe('Selected Details Tabs - Params present', () => {
        it('renders Zone as first tab and is also the initially selected on first render in Zone View/Tab', async () => {
            mockedUsePathParams.mockReturnValue({
                tagId: '1',
                ruleId: undefined,
                memberId: undefined,
                tagTypeDisplay: 'Zone',
            } as any);
            render(<SelectedDetailsTabsWrapper />);

            const allTabs = await screen.findAllByRole('tab');
            const firstTab = allTabs[0];

            expect(firstTab).toHaveTextContent(/zone/i);
            expect(firstTab).toHaveAttribute('aria-selected', 'true');
            expect(firstTab).toHaveAttribute('data-state', 'active');
        });
        it('renders Label as first tab and is also the initially selected on first render in Labels View/Tab', async () => {
            mockedUsePathParams.mockReturnValue({
                tagId: '2',
                ruleId: undefined,
                memberId: undefined,
                tagTypeDisplay: 'Label',
            } as any);
            render(<SelectedDetailsTabsWrapper />);

            const allTabs = await screen.findAllByRole('tab');
            const firstTab = allTabs[0];

            expect(firstTab).toHaveTextContent(/label/i);
            expect(firstTab).toHaveAttribute('aria-selected', 'true');
            expect(firstTab).toHaveAttribute('data-state', 'active');
        });
        it('renders enabled Zone tab and disabled Rule and Object tabs', () => {
            mockedUsePathParams.mockReturnValue({
                tagId: '1',
                ruleId: undefined,
                memberId: undefined,
                tagTypeDisplay: 'Zone',
            } as any);

            render(<SelectedDetailsTabsWrapper />);

            const zoneTab = screen.getByRole('tab', { name: /zone/i });
            const ruleTab = screen.getByRole('tab', { name: /rule/i });
            const objectTab = screen.getByRole('tab', { name: /object/i });

            expect(zoneTab).toBeEnabled();
            expect(ruleTab).toBeDisabled();
            expect(objectTab).toBeDisabled();
        });

        it('renders enabled Zone and Rule tabs and disabled Object tab', () => {
            mockedUsePathParams.mockReturnValue({
                tagId: '1',
                ruleId: '22',
                memberId: undefined,
                tagTypeDisplay: 'Zone',
            } as any);

            render(<SelectedDetailsTabsWrapper />);

            const zoneTab = screen.getByRole('tab', { name: /zone/i });
            const ruleTab = screen.getByRole('tab', { name: /rule/i });
            const objectTab = screen.getByRole('tab', { name: /object/i });

            expect(zoneTab).toBeEnabled();
            expect(ruleTab).toBeEnabled();
            expect(objectTab).toBeDisabled();
        });

        it('renders enabled Zone and Object tabs and disabled Rule tab', () => {
            mockedUsePathParams.mockReturnValue({
                tagId: '1',
                ruleId: undefined,
                memberId: '22',
                tagTypeDisplay: 'Zone',
            } as any);

            render(<SelectedDetailsTabsWrapper />);

            const zoneTab = screen.getByRole('tab', { name: /zone/i });
            const ruleTab = screen.getByRole('tab', { name: /rule/i });
            const objectTab = screen.getByRole('tab', { name: /object/i });

            expect(zoneTab).toBeEnabled();
            expect(ruleTab).toBeDisabled();
            expect(objectTab).toBeEnabled();
        });
        it('renders enabled Label, Rule and Object tabs', () => {
            mockedUsePathParams.mockReturnValue({
                tagId: '2',
                ruleId: '2',
                memberId: '22',
                tagTypeDisplay: 'Label',
            } as any);

            render(<SelectedDetailsTabsWrapper />);

            const labelTab = screen.getByRole('tab', { name: /label/i });
            const ruleTab = screen.getByRole('tab', { name: /rule/i });
            const objectTab = screen.getByRole('tab', { name: /object/i });

            expect(labelTab).toBeEnabled();
            expect(ruleTab).toBeEnabled();
            expect(objectTab).toBeEnabled();
        });
    });
});
