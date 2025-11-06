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
import { useParams } from 'react-router-dom';
import { zoneHandlers } from '../../../mocks';
import { render, screen } from '../../../test-utils';
import { PZEditButton } from './PZEditButton';

const server = setupServer(...zoneHandlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useParams: vi.fn(),
    };
});

describe('Zone Management', async () => {
    it('renders "Edit Zone" button when on Zones page', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '3', labelId: undefined, selectorId: undefined });
        render(<PZEditButton showEditButton={true} />);
        expect(await screen.findByText(/Edit Zone/i)).toBeInTheDocument();
    });

    it('renders "Edit Label" button when on Labels page', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: undefined, labelId: '1', selectorId: undefined });
        render(<PZEditButton showEditButton={true} />);
        expect(await screen.findByText(/Edit Label/i)).toBeInTheDocument();
    });

    it('renders "Edit Selector" button when rule is selected', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: undefined, labelId: undefined, selectorId: '1' });
        render(<PZEditButton showEditButton={true} />);
        expect(await screen.findByText(/Edit Rule/i)).toBeInTheDocument();
    });
});
