// Copyright 2023 Specter Ops, Inc.
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

import EntityInfoContent from './EntityInfoContent';
import { EntityInfoPanelContextProvider } from './EntityInfoPanelContextProvider';
import { render, screen, waitForElementToBeRemoved } from 'src/test-utils';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { AzureNodeKind } from 'bh-shared-ui';

const server = setupServer(
    rest.get('/api/v2/azure/roles', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    kind: 'AZRole',
                    props: {},
                    active_assignments: 0,
                    pim_assignments: 0,
                },
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('EntityInfoContent', () => {
    it('AZRole information panel will not display a section for PIM Assignments', async () => {
        const testId = '1';

        render(
            <EntityInfoPanelContextProvider>
                <EntityInfoContent id={testId} nodeType={AzureNodeKind.Role} />
            </EntityInfoPanelContextProvider>
        );
        await waitForElementToBeRemoved(() => screen.getByText('Loading...'));
        expect(screen.queryByText('PIM Assignments')).not.toBeInTheDocument();
    });
});
