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
import { usePZPathParams } from '../../../hooks';
import { zoneHandlers } from '../../../mocks';
import { render, screen } from '../../../test-utils';
import { SelectedDetailsTabs } from './SelectedDetailsTabs';

const server = setupServer(...zoneHandlers);

vi.mock('../../../hooks/usePZParams/');
const mockedUsePathParams = vi.mocked(usePZPathParams);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Selected Details Tabs', async () => {
    it('renders Zone as first tab and is also the initially selected on first render in Zone View/Tab', async () => {
        mockedUsePathParams.mockReturnValue({
            tagId: '1',
            ruleId: undefined,
            memberId: undefined,
            tagTypeDisplay: 'Zone',
        } as any);
        render(<SelectedDetailsTabs />);

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
        render(<SelectedDetailsTabs />);

        const allTabs = await screen.findAllByRole('tab');
        const firstTab = allTabs[0];

        expect(firstTab).toHaveTextContent(/label/i);
        expect(firstTab).toHaveAttribute('aria-selected', 'true');
        expect(firstTab).toHaveAttribute('data-state', 'active');
    });
    it('disables Rule and Object tabs if no ruleId and memberId', async () => {
        mockedUsePathParams.mockReturnValue({
            tagId: '1',
            ruleId: undefined,
            memberId: undefined,
            tagTypeDisplay: 'Zone',
        } as any);
        render(<SelectedDetailsTabs />);

        const allTabs = await screen.findAllByRole('tab');
        const ruleTab = allTabs[1];
        const objectTab = allTabs[2];

        expect(ruleTab).toHaveTextContent(/rule/i);
        expect(ruleTab).toHaveAttribute('data-disabled');
        expect(objectTab).toHaveTextContent(/object/i);
        expect(objectTab).toHaveAttribute('data-disabled');
    });
    it('switches tab to active/selected when clicked', async () => {
        mockedUsePathParams.mockReturnValue({
            tagId: '1',
            ruleId: '3',
            memberId: undefined,
            tagTypeDisplay: 'Zone',
        } as any);
        render(<SelectedDetailsTabs />);

        const allTabs = await screen.findAllByRole('tab');
        const zoneTab = allTabs[0];
        const ruleTab = allTabs[1];

        expect(ruleTab).toHaveTextContent(/rule/i);
        expect(ruleTab).not.toHaveAttribute('data-disabled');
        expect(ruleTab).toHaveAttribute('data-state', 'active');

        const user = userEvent.setup();
        await user.click(zoneTab);

        expect(zoneTab).toHaveAttribute('data-state', 'active');
    });
});
