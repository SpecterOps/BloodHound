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

import { setupServer } from 'msw/node';
import { Route, Routes } from 'react-router-dom';
import { zoneHandlers } from '../../../mocks';
import { detailsPath, privilegeZonesPath, zonesPath } from '../../../routes';
import { render } from '../../../test-utils';
import { SelectedDetailsTabContent } from './SelectedDetailsTabContent';
import { detailsTabOptions } from './utils';

const server = setupServer(...zoneHandlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// SPDX-License-Identifier: Apache-2.0
describe('Selected Details Tab Content', async () => {
    it('renders the Zone Tab content when Zone tab is chosen', async () => {
        render(
            <Routes>
                <Route
                    path={`/${privilegeZonesPath}/${zonesPath}/:zoneId/${detailsPath}`}
                    element={<SelectedDetailsTabContent currentDetailsTab={detailsTabOptions[0]} />}
                />
            </Routes>,
            { route: `/${privilegeZonesPath}/${zonesPath}/1/${detailsPath}` }
        );
    });
    it.skip('renders the Rule Tab content when Rule tab is chosen', () => {});
    it.skip('renders the Cypher Rules Panel when clicking the Rule Tab', () => {});
    it.skip('renders the Object Tab content when Object tab is chosen', () => {});
});
