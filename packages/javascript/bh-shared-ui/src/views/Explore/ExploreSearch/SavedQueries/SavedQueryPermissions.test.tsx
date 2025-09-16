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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, waitFor } from '../../../../test-utils';
import SavedQueryPermissions from './SavedQueryPermissions';

const testUsers = [
    {
        id: 'e504b63c-83c2-4b51-bfa4-84be18f02757',
        email_address: 'email1@email.com',
        first_name: 'Bill S',
        last_name: 'Preston',
    },
    {
        id: '09fda5c4-f0dd-4e51-8ba9-77e65300cf01',
        email_address: 'email2@email.com',
        first_name: 'Ted Theodore',
        last_name: 'Logan',
    },
    {
        id: '329a55d6-686e-4f39-afd4-8e5d528094e9',
        email_address: 'spam@example.com',
        first_name: 'BHE',
        last_name: 'Dev',
    },
];

const testSelf = {
    id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
};

const handlers = [
    rest.get('/api/v2/bloodhound-users-minimal', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    users: testUsers,
                },
            })
        );
    }),
    rest.get(`/api/v2/self`, async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: testSelf,
            })
        );
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('SavedQueryPermissions', () => {
    it('should render an SavedQueryPermissions component', async () => {
        const testSetIsPublic = vi.fn();
        const testSetSharedIds = vi.fn();
        const testSharedIds: string[] = [];
        const testIsPublic = false;

        render(
            <SavedQueryPermissions
                sharedIds={testSharedIds}
                isPublic={testIsPublic}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );

        // Empty State
        expect(screen.getByText(/loading/i)).toBeInTheDocument();

        // Table Header Rendered
        const nestedElement = await waitFor(() => screen.getByText(/name/i));
        expect(nestedElement).toBeInTheDocument();

        expect(screen.getByRole('checkbox')).toBeInTheDocument();
    });

    it('should should reflect the checked state for public query', async () => {
        const testSetIsPublic = vi.fn();
        const testSetSharedIds = vi.fn();
        const testSharedIds: string[] = [];
        const testIsPublic = true;
        render(
            <SavedQueryPermissions
                sharedIds={testSharedIds}
                isPublic={testIsPublic}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );
        const nestedElement = await waitFor(() => screen.getByText(/name/i));
        expect(nestedElement).toBeInTheDocument();
        const setPublicCheckbox = screen.getByTestId('public-query');
        expect(setPublicCheckbox).toBeInTheDocument();
        expect(setPublicCheckbox).toBeChecked();
    });

    it('should should reflect the NOT checked state for public query', async () => {
        const testSetIsPublic = vi.fn();
        const testSetSharedIds = vi.fn();
        const testSharedIds: string[] = [];
        const testIsPublic = false;
        render(
            <SavedQueryPermissions
                sharedIds={testSharedIds}
                isPublic={testIsPublic}
                setSharedIds={testSetSharedIds}
                setIsPublic={testSetIsPublic}
            />
        );
        const nestedElement = await waitFor(() => screen.getByText(/name/i));
        expect(nestedElement).toBeInTheDocument();
        const setPublicCheckbox = screen.getByTestId('public-query');
        expect(setPublicCheckbox).toBeInTheDocument();
        expect(setPublicCheckbox).not.toBeChecked();
    });
});
