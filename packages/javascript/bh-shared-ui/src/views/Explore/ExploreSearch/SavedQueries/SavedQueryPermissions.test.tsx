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
        principal_name: 'Ted Theodore Logan - user',
        id: '8f98c54d-ac4b-464d-a391-d8d5d4b3fe8c',
    },
    {
        principal_name: 'Socrates - read-only',
        id: 'cd28625d-5a09-4312-b84a-72a95881387a',
    },
    {
        principal_name: 'Beethoven - upload-only',
        id: '46cd933d-b556-4fd6-a7a5-c0c7a44ea11e',
    },
    {
        principal_name: 'Joan of Arc - power-user',
        id: 'b1ebcec3-97bd-4660-8a88-b1895cbf4859',
    },
    {
        principal_name: 'Bill S Preston - admin',
        id: '0bf8e936-c70b-4ddc-ad8a-98a3afbf6393',
    },
];

const testSelf = {
    id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
};

const handlers = [
    rest.get('/api/v2/bloodhound-users', (req, res, ctx) => {
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
