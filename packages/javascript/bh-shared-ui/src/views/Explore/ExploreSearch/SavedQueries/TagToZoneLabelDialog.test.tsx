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

import TagToZoneLabelDialog from './TagToZoneLabelDialog';

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { vi } from 'vitest';
import { render, screen } from '../../../../test-utils';

const testZones = [
    {
        id: 34,
        type: 1,
        kind_id: 199,
        name: 'Tier One',
        description: 'T1',
        created_at: '2025-08-04T22:01:58.504711Z',
        created_by: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
        updated_at: '2025-08-04T22:01:58.504711Z',
        updated_by: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
        deleted_at: null,
        deleted_by: null,
        position: 2,
        require_certify: false,
        analysis_enabled: false,
        counts: {
            selectors: 0,
            members: 0,
        },
    },
    {
        id: 1,
        type: 1,
        kind_id: 175,
        name: 'Tier Zero',
        description: 'Tier Zero',
        created_at: '2025-08-04T21:23:41.015481Z',
        created_by: 'SYSTEM',
        updated_at: '2025-08-04T21:23:41.015481Z',
        updated_by: 'SYSTEM',
        deleted_at: null,
        deleted_by: null,
        position: 1,
        require_certify: false,
        analysis_enabled: true,
        counts: {
            selectors: 17,
            members: 177,
        },
    },
];

const handlers = [
    rest.get('/api/v2/asset-groups', (req, res, ctx) => {
        return res(
            ctx.json({
                data: testZones,
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags', (req, res, ctx) => {
        return res(
            ctx.json({
                data: {},
            })
        );
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('TagToZoneLabelDialog', () => {
    it('should render a TagToZoneLabelDialog component', () => {
        const testHandleSetOpen = vi.fn();

        render(
            <TagToZoneLabelDialog
                dialogOpen={true}
                setDialogOpen={testHandleSetOpen}
                isLabel={false}
                selectedQuery={undefined}
                cypherQuery={''}
            />
        );

        expect(screen.getByText(/Tag Results to Zone/i)).toBeInTheDocument();
        expect(screen.queryByText(/Tag Results to Label/i)).not.toBeInTheDocument();
    });

    it('should render tag to Label content', () => {
        const testHandleSetOpen = vi.fn();

        render(
            <TagToZoneLabelDialog
                dialogOpen={true}
                setDialogOpen={testHandleSetOpen}
                isLabel={true}
                selectedQuery={undefined}
                cypherQuery={''}
            />
        );

        expect(screen.getByText(/Tag Results to Label/i)).toBeInTheDocument();
        expect(screen.queryByText(/Tag Results to Zone/i)).not.toBeInTheDocument();
    });

    it('should handle close event', async () => {
        const user = userEvent.setup();
        const testHandleSetOpen = vi.fn();

        render(
            <TagToZoneLabelDialog
                dialogOpen={true}
                setDialogOpen={testHandleSetOpen}
                isLabel={true}
                selectedQuery={undefined}
                cypherQuery={''}
            />
        );

        await user.click(screen.getByRole('button', { name: /cancel/i }));
        expect(testHandleSetOpen).toHaveBeenCalled();
        expect(testHandleSetOpen).toHaveBeenCalledTimes(1);
    });

    it('should have a disabled continue button', async () => {
        const testHandleSetOpen = vi.fn();
        render(
            <TagToZoneLabelDialog
                dialogOpen={true}
                setDialogOpen={testHandleSetOpen}
                isLabel={true}
                selectedQuery={undefined}
                cypherQuery={''}
            />
        );
        const testContinue = screen.getByRole('button', { name: /continue/i });
        expect(testContinue).toBeInTheDocument();
        expect(testContinue).toBeDisabled();
    });
});
