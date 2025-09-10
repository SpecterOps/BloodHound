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
import { ConfigurationKey } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { Route, Routes, useParams } from 'react-router-dom';
import TagForm from '.';
import { act, fireEvent, longWait, render, screen, waitFor, within } from '../../../../test-utils';

const testTierZero = {
    id: 1,
    type: 1,
    kind_id: 173,
    name: 'Tier Zero',
    description: 'Tier Zero Description',
    created_at: '2025-04-15T21:02:26.504736Z',
    created_by: 'SYSTEM',
    updated_at: '2025-04-15T21:02:26.504736Z',
    updated_by: 'SYSTEM',
    deleted_at: null,
    deleted_by: null,
    position: 1,
    require_certify: false,
    analysis_enabled: true,
};

const testOwned = {
    id: 2,
    type: 3,
    kind_id: 173,
    name: 'Owned',
    description: 'Owned Description',
    created_at: '2025-04-15T21:02:26.504736Z',
    created_by: 'SYSTEM',
    updated_at: '2025-04-15T21:02:26.504736Z',
    updated_by: 'SYSTEM',
    deleted_at: null,
    deleted_by: null,
    position: null,
    require_certify: false,
    analysis_enabled: false,
};

const handlers = [
    rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(ctx.json({ data: { tags: [] } }));
    }),
    rest.post('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(ctx.json({ data: { tag: { id: 777 } } }));
    }),
    rest.patch('/api/v2/asset-group-tags/:tagId', async (_, res, ctx) => {
        return res(ctx.json({ data: { tag: { id: 777 } } }));
    }),
    rest.delete('/api/v2/asset-group-tags/:tagId', async (_, res, ctx) => {
        return res(ctx.status(500, 'get rekt'));
    }),
    rest.get('/api/v2/asset-group-tags/1', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: testTierZero,
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags/2', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: testOwned,
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags/3', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: { ...testOwned, name: 'myTestLabel', id: 3, type: 2 },
            })
        );
    }),
    rest.get('/api/v2/config', async (_, res, ctx) => {
        return res(ctx.json(configResponse));
    }),
    rest.get('/api/v2/features', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        key: 'tier_management_engine',
                        enabled: true,
                    },
                ],
            })
        );
    }),
];

const configResponse = {
    data: [
        {
            key: ConfigurationKey.Tiering,
            value: { multi_tier_analysis_enabled: true, tier_limit: 1, label_limit: 0 },
        },
    ],
};

const server = setupServer(...handlers);

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

const mockNavigate = vi.fn();
vi.mock('../../../../utils', async () => {
    const actual = await vi.importActual('../../../../utils');
    return {
        ...actual,
        useAppNavigate: () => mockNavigate,
    };
});

const mockAddNotification = vi.fn();
vi.mock('../../../../providers', async () => {
    const actual = await vi.importActual('../../../../providers');
    return {
        ...actual,
        addNotification: () => mockAddNotification,
    };
});

describe('Tag Form', () => {
    const user = userEvent.setup();
    // todo links to Details
    const createNewZonePath = '/privilege-zones/zone/save/';
    const createNewLabelPath = '/privilege-zones/label/save/';
    const editExistingZonePath = '/privilege-zones/zone/1/save';
    const editExistingLabelPath = '/privilege-zones/label/2/save';
    const deletionTestsPath = '/privilege-zones/label/3/save';

    it('renders the form for creating a new zone', async () => {
        // Because there is no id path parameter in the url, the form is a create form
        // This means that none of the input fields should have any value aside from default values

        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });

        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configResponse));
            })
        );

        render(
            <Routes>
                <Route path={createNewZonePath} element={<TagForm />} />
            </Routes>,
            { route: createNewZonePath }
        );

        expect(await screen.findByText('Create new Zone')).toBeInTheDocument();

        const nameInput = screen.getByLabelText('Name');
        expect(nameInput).toBeInTheDocument();
        expect(nameInput).toHaveValue('');

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        expect(descriptionInput).toHaveValue('');

        // The delete button should not render when creating a new selector because it doesn't exist yet
        expect(screen.queryByRole('button', { name: /Delete/ })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Define Selector/ })).toBeInTheDocument();
        expect(screen.queryByText(/Enable Analysis/i)).not.toBeInTheDocument();
    });

    it('renders the form for creating a new label', async () => {
        // Because there is no id path parameter in the url, the form is a create form
        // This means that none of the input fields should have any value aside from default values

        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });

        render(
            <Routes>
                <Route path={createNewLabelPath} element={<TagForm />} />
            </Routes>,
            { route: createNewLabelPath }
        );

        expect(await screen.findByText('Create new Label')).toBeInTheDocument();

        const nameInput = screen.getByLabelText('Name');
        expect(nameInput).toBeInTheDocument();
        expect(nameInput).toHaveValue('');

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        expect(descriptionInput).toHaveValue('');

        // The delete button should not render when creating a new selector because it doesn't exist yet
        expect(screen.queryByRole('button', { name: /Delete/ })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Define Selector/ })).toBeInTheDocument();
        expect(screen.queryByText(/Enable Analysis/i)).not.toBeInTheDocument();
    });

    it('does not render the analysis toggle when multi tier analysis enabled is false', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '2', labelId: undefined });

        const configResponse = {
            data: [
                {
                    key: ConfigurationKey.Tiering,
                    value: { multi_tier_analysis_enabled: false, tier_limit: 1, label_limit: 0 },
                },
            ],
        };

        server.use(
            rest.get('/api/v2/config', async (_, res, ctx) => {
                return res(ctx.json(configResponse));
            })
        );

        render(
            <Routes>
                <Route path={editExistingZonePath} element={<TagForm />} />
            </Routes>,
            { route: editExistingZonePath }
        );

        expect(await screen.findByText('Edit Zone Details')).toBeInTheDocument();
        expect(screen.queryByText(/Enable Analysis/i)).not.toBeInTheDocument();
    });

    it('renders the analysis toggle when multi tier analysis enabled is true and when editing an existing zone', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '2', labelId: undefined });

        render(
            <Routes>
                <Route path={editExistingZonePath} element={<TagForm />} />
            </Routes>,
            { route: editExistingZonePath }
        );

        expect(await screen.findByText('Edit Zone Details')).toBeInTheDocument();
        expect(await screen.findByText(/Enable Analysis/i)).toBeInTheDocument();
    });

    it('renders the form for editing an existing zone', async () => {
        // This url has the zone id of 1 in the path
        // and so this zone's data is filled into the form for the user to edit

        vi.mocked(useParams).mockReturnValue({ zoneId: '1', labelId: undefined });

        render(
            <Routes>
                <Route path={editExistingZonePath} element={<TagForm />} />
            </Routes>,
            { route: editExistingZonePath }
        );

        expect(await screen.findByText('Edit Zone Details')).toBeInTheDocument();

        const nameInput = await screen.findByLabelText('Name');
        expect(nameInput).toBeInTheDocument();
        longWait(() => {
            expect(nameInput).toHaveValue('Tier Zero');
        });

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        longWait(() => {
            expect(descriptionInput).toHaveValue('Tier Zero Description');
        });

        // The delete button should not render when editing T0
        expect(screen.queryByRole('button', { name: /Delete/ })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Save/ })).toBeInTheDocument();
    });

    it('renders the form for editing an existing label', async () => {
        // This url has the label id of 2 in the path
        // and so this label's data is filled into the form for the user to edit

        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: '2' });

        render(
            <Routes>
                <Route path={editExistingLabelPath} element={<TagForm />} />
            </Routes>,
            { route: editExistingLabelPath }
        );

        longWait(() => {
            expect(screen.getByText('Edit Label Details')).toBeInTheDocument();
        });

        const nameInput = await screen.findByLabelText('Name');
        expect(nameInput).toBeInTheDocument();
        longWait(() => {
            expect(nameInput).toHaveValue('Owned');
        });

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        longWait(() => {
            expect(descriptionInput).toHaveValue('Owned Description');
        });

        // The delete button should not render when editing Owned
        longWait(() => {
            expect(screen.queryByRole('button', { name: /Delete/ })).not.toBeInTheDocument();
        });
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Save/ })).toBeInTheDocument();
        expect(screen.queryByTestId('zone-management_save_tag-form_analysis-enabled-switch')).not.toBeInTheDocument();
    });

    test('clicking cancel on the form takes the user back to the page the user was on previously', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: '2' });
        render(<TagForm />, { route: createNewLabelPath });

        await act(async () => {
            fireEvent.click(await screen.findByTestId('zone-management_save_tag-form_cancel-button'));
        });

        await waitFor(() => {
            expect(mockNavigate).toHaveBeenCalledWith(-1);
        });
    });

    test('a name value is required to submit the form', async () => {
        render(<TagForm />, { route: createNewLabelPath });

        await user.click(await screen.findByRole('button', { name: /Save/ }));

        longWait(() => {
            expect(screen.getByText('Please provide a name for the label')).toBeInTheDocument();
        });
    });

    it('validates that the name input is under the max length', async () => {
        render(<TagForm />, { route: createNewLabelPath });

        const nameInput = await screen.findByLabelText('Name');

        await user.click(nameInput);
        await user.paste('f'.repeat(251));

        await user.click(await screen.findByRole('button', { name: /Save/ }));

        longWait(() => {
            expect(
                screen.getByText(`Name cannot exceed 250 characters. Please provide a shorter name`)
            ).toBeInTheDocument();
        });
    });

    test('filling in the name value allows updating the selector and navigates back to the details page', async () => {
        render(
            <Routes>
                <Route path={'/'} element={<TagForm />} />
                <Route path={createNewZonePath} element={<TagForm />} />
            </Routes>,
            { route: createNewZonePath }
        );

        const nameInput = await screen.findByLabelText('Name');

        await user.click(nameInput);
        await user.paste('foo');

        await user.click(await screen.findByRole('button', { name: /Save/ }));

        expect(screen.queryByText('Please provide a name for the zone')).not.toBeInTheDocument();

        await waitFor(() => {
            expect(mockNavigate).toBeCalled();
        });
    });

    it('handles creating a new zone', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });

        render(
            <Routes>
                <Route path={'/'} element={<TagForm />} />
                <Route path={createNewZonePath} element={<TagForm />} />
            </Routes>,
            { route: createNewZonePath }
        );

        const nameInput = await screen.findByLabelText('Name');

        await user.click(nameInput);
        await user.paste('foo');

        await user.click(await screen.findByRole('button', { name: /Define Selector/ }));

        longWait(() => {
            expect(mockNavigate).toBeCalled();
            expect(mockAddNotification).toBeCalled();
        });
    });

    it('handles creating a new label', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });

        render(
            <Routes>
                <Route path={'/'} element={<TagForm />} />
                <Route path={createNewLabelPath} element={<TagForm />} />
            </Routes>,
            { route: createNewLabelPath }
        );

        const nameInput = await screen.findByLabelText('Name');

        await user.click(nameInput);
        await user.paste('foo');

        await user.click(await screen.findByRole('button', { name: /Define Selector/ }));

        longWait(() => {
            expect(mockNavigate).toBeCalled();
            expect(mockAddNotification).toBeCalled();
        });
    });

    it('disables the confirm button until user types required text', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: '3' });

        render(<TagForm />);

        const deleteButton = await screen.findByTestId('zone-management_save_tag-form_delete-button');

        await act(async () => {
            fireEvent.click(deleteButton);
        });

        await waitFor(() => {
            expect(screen.getByRole('dialog')).toBeInTheDocument();
        });

        const confirmButton = screen.getByRole('button', { name: 'Confirm' });
        expect(confirmButton).toBeDisabled();

        const textField = screen.getByTestId('confirmation-dialog_challenge-text');
        await user.type(textField, 'incorrect text');

        expect(confirmButton).toBeDisabled();

        await user.clear(textField);
        await user.type(textField, 'delete this label');

        expect(confirmButton).not.toHaveAttribute('disabled', true);
    });

    it('opens and closes the dialog with the cancel button', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: '3' });

        render(<TagForm />);

        const deleteButton = await screen.findByTestId('zone-management_save_tag-form_delete-button');

        await act(async () => {
            fireEvent.click(deleteButton);
        });

        await waitFor(() => {
            expect(screen.getByRole('dialog')).toBeInTheDocument();
        });

        const closeButton = within(screen.getByRole('dialog')).getByRole('button', { name: /cancel/i });

        await act(async () => {
            await user.click(closeButton);
        });

        await waitFor(() => {
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        });
    });

    it('open and closes dialog with confirm button after user inputs required text', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: '3' });
        console.error = vi.fn();

        render(
            <Routes>
                <Route path={deletionTestsPath} element={<TagForm />} />
            </Routes>,
            { route: deletionTestsPath }
        );

        const deleteButton = await screen.findByRole('button', { name: /Delete Label/i });
        await act(async () => {
            fireEvent.click(deleteButton);
        });

        await waitFor(() => {
            expect(screen.getByRole('dialog')).toBeInTheDocument();
        });

        const confirmButton = screen.getByRole('button', { name: /confirm/i });
        expect(confirmButton).toBeDisabled();

        const textField = screen.getByTestId('confirmation-dialog_challenge-text');
        await user.type(textField, 'Delete this label');

        expect(confirmButton).not.toBeDisabled();
        await user.click(confirmButton);

        waitFor(() => {
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        });
    });
});
