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
    it('render Zones as first tab and is also the initial selected on first render in Zone View/Tab', async () => {
        mockedUsePathParams.mockReturnValue({
            tagId: '1',
            ruleId: undefined,
            memberId: undefined,
            tagTypeDisplay: 'Zone',
        } as any);
        render(<SelectedDetailsTabs />);
        const firstTabTitle = await screen.findByRole('tab', { name: /zone/i });
        expect(firstTabTitle).toBeInTheDocument();
    });
    it('render Label as first tab and is also the initial selected on first render in Labels View/Tab', async () => {
        mockedUsePathParams.mockReturnValue({
            tagId: '2',
            ruleId: undefined,
            memberId: undefined,
            tagTypeDisplay: 'Label',
        } as any);
        render(<SelectedDetailsTabs />);
        const firstTabTitle = await screen.findByRole('tab', { name: /label/i });
        expect(firstTabTitle).toBeInTheDocument();
    });
    // Other Tabs are disabled when no path params
    // When clicking on Tab changes tab
});
