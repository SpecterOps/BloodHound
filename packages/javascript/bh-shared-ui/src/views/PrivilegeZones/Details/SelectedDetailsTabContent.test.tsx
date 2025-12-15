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
import { zoneHandlers } from '../../../mocks';
import { render, screen } from '../../../test-utils';
import { SelectedDetailsTabContent } from './SelectedDetailsTabContent';
import { detailsTabOptions } from './utils';

const server = setupServer(...zoneHandlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Selected Details Tab Content', async () => {
    it('renders the Zone/Labels Tab content when first tab is chosen', async () => {
        render(<SelectedDetailsTabContent currentDetailsTab={detailsTabOptions[0]} tagId='1' />);
        const zoneTitle = await screen.findByText(/tier-0/i); // can find the structure of title in mocks/factories/privilegeZones
        expect(zoneTitle).toBeInTheDocument();
    });
    it('renders the Rule Tab content when Rule tab is chosen', async () => {
        render(<SelectedDetailsTabContent currentDetailsTab={detailsTabOptions[1]} tagId='1' ruleId='2' />);
        const ruleTitle = await screen.findByText(/tier-0-rule-2/i); // can find the structure of title in mocks/factories/privilegeZones
        expect(ruleTitle).toBeInTheDocument();
    });
    it.skip('renders the Cypher Rules Panel when clicking the Rule Tab', () => {});
    it('renders the Object Tab content when Object tab is chosen', async () => {
        render(
            <SelectedDetailsTabContent currentDetailsTab={detailsTabOptions[2]} tagId='1' ruleId='2' memberId='1' />
        );
        const entityInfoPanel = await screen.findByTestId('selected-details-object-panel'); // had to add wrapper test id as panel components has explore references
        expect(entityInfoPanel).toBeInTheDocument();
    });
});
