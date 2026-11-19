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
import { Location, Route, Routes, useLocation, useParams } from 'react-router-dom';
import TagForm from '.';
import { privilegeZonesKeys, usePatchRule } from '../../../../hooks';
import * as notificationActions from '../../../../providers/NotificationProvider/actions';
import { labelsPath, privilegeZonesPath, savePath, zonesPath } from '../../../../routes';
import { act, fireEvent, render, renderHook, screen, waitFor, within } from '../../../../test-utils';
import { setUpQueryClient } from '../../../../testing';
import { apiClient } from '../../../../utils';

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
    glyph: 'lightbulb',
    position: 1,
    require_certify: true,
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
    glyph: null,
    position: null,
    require_certify: false,
    analysis_enabled: false,
};

const handlers = [
    rest.get('/api/v2/asset-group-tags/:tagId/selectors', async (_, res, ctx) => {
        return res(ctx.json({ data: { selectors: [] }, count: 0 }));
    }),
    rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(ctx.json({ data: { tags: [testTierZero, testOwned] } }));
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
                data: { tag: testTierZero },
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags/2', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: { tag: testOwned },
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags/3', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: { tag: { ...testOwned, name: 'myTestLabel', id: 3, type: 2 } },
            })
        );
    }),
    rest.get('/api/v2/asset-group-tags/4', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: { tag: { ...testTierZero, name: 'other zone', id: 4, type: 1, position: 2 } },
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
        useLocation: vi.fn(),
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

describe('Tag Form', () => {
    const user = userEvent.setup();
    const createNewZonePath = `/${privilegeZonesPath}/${zonesPath}/${savePath}/`;
    const createNewLabelPath = `/${privilegeZonesPath}/${labelsPath}/${savePath}/`;
    const editHighestPrivilegeZonePath = `/${privilegeZonesPath}/${zonesPath}/1/${savePath}`;
    const editOtherZonePath = `/${privilegeZonesPath}/${zonesPath}/4/${savePath}`;
    const editExistingLabelPath = `/${privilegeZonesPath}/${labelsPath}/2/${savePath}`;
    const deletionTestsPath = `/${privilegeZonesPath}/${labelsPath}/3/${savePath}`;

    const renderLabelForm = (queryClient = setUpQueryClient([])) => {
        vi.mocked(useParams).mockReturnValue({ labelId: '2' });
        vi.mocked(useLocation).mockReturnValue({ pathname: editExistingLabelPath } as Location);
        return render(
            <Routes>
                <Route path={editExistingLabelPath} element={<TagForm />} />
            </Routes>,
            { route: editExistingLabelPath, queryClient }
        );
    };

    it.each(['enabled', 'disabled', 'mixed'] as const)(
        'updates every rule across pages when toggling a label with %s rules',
        async (initialStatus) => {
            mockNavigate.mockClear();
            const rules = Array.from({ length: 27 }, (_, index) => ({
                id: index + 1,
                allow_disable: true,
                disabled_at:
                    initialStatus === 'disabled' || (initialStatus === 'mixed' && index < 26)
                        ? '2026-01-01T00:00:00Z'
                        : null,
            }));
            const updatedRuleIds: number[] = [];
            const tagUpdates = vi.fn();
            server.use(
                rest.get('/api/v2/asset-group-tags/2/selectors', (request, res, ctx) => {
                    const skip = Number(request.url.searchParams.get('skip'));
                    const limit = Number(request.url.searchParams.get('limit'));
                    return res(ctx.json({ data: { selectors: rules.slice(skip, skip + limit) }, count: rules.length }));
                }),
                rest.patch('/api/v2/asset-group-tags/2/selectors/:ruleId', async (request, res, ctx) => {
                    const { disabled } = await request.json();
                    const rule = rules.find((rule) => rule.id === Number(request.params.ruleId))!;
                    rule.disabled_at = disabled ? '2026-01-01T00:00:00Z' : null;
                    updatedRuleIds.push(rule.id);
                    return res(ctx.json({ data: { selector: rule } }));
                }),
                rest.patch('/api/v2/asset-group-tags/2', (_, res, ctx) => {
                    tagUpdates();
                    return res(ctx.json({ data: { tag: testOwned } }));
                })
            );

            renderLabelForm();
            const toggle = await screen.findByLabelText('Enable Rules');
            await waitFor(() => expect(toggle).toBeEnabled());
            expect(toggle).toHaveAttribute('aria-checked', String(initialStatus !== 'disabled'));

            await user.click(toggle);
            await user.click(screen.getByRole('button', { name: 'Save Edits' }));

            await waitFor(() => expect(mockNavigate).toHaveBeenCalled());
            expect(updatedRuleIds).toEqual(initialStatus === 'mixed' ? [27] : rules.map((rule) => rule.id));
            expect(rules.every((rule) => (rule.disabled_at === null) === (initialStatus === 'disabled'))).toBe(true);
            expect(tagUpdates).not.toHaveBeenCalled();
        }
    );

    it.each([
        { initiallyEnabled: false, disabled: false },
        { initiallyEnabled: true, disabled: true },
    ])(
        'keeps the toggle on with one enabled rule after updating a rule to disabled=$disabled',
        async ({ initiallyEnabled, disabled }) => {
            const rules = [
                { id: 1, allow_disable: true, disabled_at: initiallyEnabled ? null : '2026-01-01T00:00:00Z' },
                { id: 2, allow_disable: true, disabled_at: initiallyEnabled ? null : '2026-01-01T00:00:00Z' },
            ];
            const updatedRuleIds: number[] = [];
            server.use(
                rest.get('/api/v2/asset-group-tags/2/selectors', (_, res, ctx) => {
                    return res(ctx.json({ data: { selectors: rules }, count: rules.length }));
                }),
                rest.patch('/api/v2/asset-group-tags/2/selectors/:ruleId', async (request, res, ctx) => {
                    const { disabled } = await request.json();
                    const rule = rules.find((rule) => rule.id === Number(request.params.ruleId))!;
                    rule.disabled_at = disabled ? '2026-01-01T00:00:00Z' : null;
                    updatedRuleIds.push(rule.id);
                    return res(ctx.json({ data: { selector: rule } }));
                })
            );
            const queryClient = setUpQueryClient([]);
            renderLabelForm(queryClient);
            const toggle = await screen.findByLabelText('Enable Rules');
            await waitFor(() => expect(toggle).toBeEnabled());
            expect(toggle).toHaveAttribute('aria-checked', String(initiallyEnabled));

            const { result } = renderHook(() => usePatchRule('2'), { queryClient });
            await act(async () => {
                await result.current.mutateAsync({ tagId: '2', ruleId: 1, updatedValues: { id: 1, disabled } });
            });

            await waitFor(() =>
                expect(queryClient.getQueryData([...privilegeZonesKeys.rulesByTag('2'), 'all'])).toEqual(
                    expect.arrayContaining([
                        expect.objectContaining({ id: 1, disabled_at: disabled ? '2026-01-01T00:00:00Z' : null }),
                    ])
                )
            );
            await waitFor(() => expect(toggle).toHaveAttribute('aria-checked', 'true'));
            expect(updatedRuleIds).toEqual([1]);
            expect(rules.filter((rule) => rule.disabled_at === null)).toHaveLength(1);
            expect(rules[1].disabled_at).toBe(initiallyEnabled ? null : '2026-01-01T00:00:00Z');
        }
    );

    it('preserves mixed rule statuses when only the label description is edited', async () => {
        const ruleUpdates = vi.fn();
        const tagUpdates = vi.fn();
        server.use(
            rest.get('/api/v2/asset-group-tags/2/selectors', (_, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            selectors: [
                                { id: 1, disabled_at: null },
                                { id: 2, disabled_at: '2026-01-01T00:00:00Z' },
                            ],
                        },
                        count: 2,
                    })
                );
            }),
            rest.patch('/api/v2/asset-group-tags/2/selectors/:ruleId', (_, res, ctx) => {
                ruleUpdates();
                return res(ctx.json({ data: {} }));
            }),
            rest.patch('/api/v2/asset-group-tags/2', async (request, res, ctx) => {
                tagUpdates(await request.json());
                return res(ctx.json({ data: { tag: testOwned } }));
            })
        );

        renderLabelForm();
        await waitFor(() => expect(screen.getByLabelText('Enable Rules')).toBeEnabled());
        const description = screen.getByLabelText('Description');
        await user.clear(description);
        await user.type(description, 'Updated description');
        await user.click(screen.getByRole('button', { name: 'Save Edits' }));

        await waitFor(() => expect(tagUpdates).toHaveBeenCalledWith({ description: 'Updated description' }));
        expect(ruleUpdates).not.toHaveBeenCalled();
    });

    it('clears the pending bulk selection when switching labels', async () => {
        server.use(
            rest.get('/api/v2/asset-group-tags/:tagId/selectors', (_, res, ctx) => {
                return res(ctx.json({ data: { selectors: [{ id: 1, disabled_at: null }] }, count: 1 }));
            })
        );
        vi.mocked(useParams).mockReturnValue({ labelId: '2' });
        vi.mocked(useLocation).mockReturnValue({ pathname: editExistingLabelPath } as Location);
        const { rerender } = render(<TagForm />, { route: editExistingLabelPath });

        const toggle = await screen.findByLabelText('Enable Rules');
        await waitFor(() => expect(toggle).toBeEnabled());
        await user.click(toggle);
        expect(toggle).toHaveAttribute('aria-checked', 'false');

        vi.mocked(useParams).mockReturnValue({ labelId: '3' });
        vi.mocked(useLocation).mockReturnValue({ pathname: deletionTestsPath } as Location);
        rerender(<TagForm />);

        await waitFor(() => expect(screen.getByPlaceholderText('Name')).toHaveValue('myTestLabel'));
        await waitFor(() => expect(screen.getByLabelText('Enable Rules')).toBeEnabled());
        expect(screen.getByLabelText('Enable Rules')).toHaveAttribute('aria-checked', 'true');
    });

    it.each([
        { enabled: true, failedCount: 1, message: 'Failed to enable 1 rule. Please try again.' },
        { enabled: true, failedCount: 2, message: 'Failed to enable 2 rules. Please try again.' },
        { enabled: false, failedCount: 1, message: 'Failed to disable 1 rule. Please try again.' },
        { enabled: false, failedCount: 2, message: 'Failed to disable 2 rules. Please try again.' },
        { enabled: false, failedCount: 3, message: 'Failed to disable 3 rules. Please try again.' },
        {
            enabled: true,
            failedCount: 1,
            message: 'Failed to enable 1 rule. Please try again.',
            updatedDescription: 'Updated description',
        },
        {
            enabled: false,
            failedCount: 1,
            message: 'Failed to disable 1 rule. Please try again.',
            updatedDescription: 'Updated description',
        },
    ])(
        'retries $failedCount failed rule updates with enabled=$enabled and description edit $updatedDescription',
        async ({ enabled, failedCount, message, updatedDescription }) => {
            mockNavigate.mockClear();
            const addNotification = vi.spyOn(notificationActions, 'addNotification');
            const tagUpdates = vi.fn();
            let savedTag = { ...testOwned };
            const rules = Array.from({ length: 3 }, (_, index) => ({
                id: index + 1,
                allow_disable: true,
                disabled_at: enabled ? ('2026-01-01T00:00:00Z' as string | null) : null,
            }));
            const attemptedRuleIds: number[] = [];
            let shouldFail = true;
            server.use(
                rest.get('/api/v2/asset-group-tags/2/selectors', (_, res, ctx) => {
                    return res(ctx.json({ data: { selectors: rules }, count: rules.length }));
                }),
                rest.patch('/api/v2/asset-group-tags/2/selectors/:ruleId', async (request, res, ctx) => {
                    const ruleId = Number(request.params.ruleId);
                    attemptedRuleIds.push(ruleId);
                    if (shouldFail && ruleId <= failedCount) return res(ctx.status(500));
                    const { disabled } = await request.json();
                    const rule = rules.find((rule) => rule.id === ruleId)!;
                    rule.disabled_at = disabled ? '2026-01-01T00:00:00Z' : null;
                    return res(ctx.json({ data: { selector: rule } }));
                }),
                rest.get('/api/v2/asset-group-tags/2', (_, res, ctx) => {
                    return res(ctx.json({ data: { tag: savedTag } }));
                }),
                rest.patch('/api/v2/asset-group-tags/2', async (request, res, ctx) => {
                    const updates = await request.json();
                    tagUpdates(updates);
                    savedTag = { ...savedTag, ...updates };
                    return res(ctx.json({ data: { tag: savedTag } }));
                })
            );

            const queryClient = setUpQueryClient([]);
            renderLabelForm(queryClient);
            const toggle = await screen.findByLabelText('Enable Rules');
            await waitFor(() => expect(toggle).toBeEnabled());
            if (updatedDescription) {
                const description = screen.getByLabelText('Description');
                await user.clear(description);
                await user.type(description, updatedDescription);
            }
            await user.click(toggle);
            await user.click(screen.getByRole('button', { name: 'Save Edits' }));

            await waitFor(() =>
                expect(addNotification).toHaveBeenCalledWith(
                    message,
                    expect.any(String),
                    expect.objectContaining({ variant: 'error' })
                )
            );
            expect(attemptedRuleIds).toEqual([1, 2, 3]);
            expect(mockNavigate).not.toHaveBeenCalled();
            expect(addNotification).not.toHaveBeenCalledWith(
                'Label was updated successfully!',
                expect.any(String),
                expect.any(Object)
            );
            expect(screen.getByRole('button', { name: 'Save Edits' })).toBeEnabled();
            if (updatedDescription) {
                expect(queryClient.getQueryData(privilegeZonesKeys.tagDetail('2'))).toMatchObject({
                    description: updatedDescription,
                });
                expect(toggle).toHaveAttribute('aria-checked', String(enabled));
            }

            shouldFail = false;
            attemptedRuleIds.length = 0;
            await user.click(screen.getByRole('button', { name: 'Save Edits' }));

            await waitFor(() => expect(mockNavigate).toHaveBeenCalled());
            expect(attemptedRuleIds).toEqual(Array.from({ length: failedCount }, (_, index) => index + 1));
            expect(rules.every((rule) => (rule.disabled_at === null) === enabled)).toBe(true);
            expect(tagUpdates.mock.calls).toEqual(updatedDescription ? [[{ description: updatedDescription }]] : []);
            addNotification.mockRestore();
        }
    );

    it('renders the form for creating a new zone', async () => {
        // Because there is no id path parameter in the url, the form is a create form
        // This means that none of the input fields should have any value aside from default values
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: createNewZonePath } as Location);

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

        const nameInput = screen.getByLabelText('Zone Information');
        expect(nameInput).toBeInTheDocument();
        expect(nameInput).toHaveValue('');

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        expect(descriptionInput).toHaveValue('');

        // The Require Certification switch should not render on BHCE
        expect(screen.queryByText(/Require Certification/i)).not.toBeInTheDocument();

        const glyphInput = await screen.findByLabelText(/Apply Custom Glyph/);
        expect(glyphInput).toBeInTheDocument();
        expect(glyphInput).toHaveValue('');

        // The delete button should not render when creating a new selector because it doesn't exist yet
        expect(screen.queryByRole('button', { name: /Delete/ })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Define Rule/ })).toBeInTheDocument();
        expect(screen.queryByText(/Enable Analysis/i)).not.toBeInTheDocument();
    });

    it('renders the form for creating a new label', async () => {
        // Because there is no id path parameter in the url, the form is a create form
        // This means that none of the input fields should have any value aside from default values

        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: createNewLabelPath } as Location);

        render(
            <Routes>
                <Route path={createNewLabelPath} element={<TagForm />} />
            </Routes>,
            { route: createNewLabelPath }
        );

        expect(await screen.findByText('Create new Label')).toBeInTheDocument();

        const nameInput = screen.getByLabelText('Label Information');
        expect(nameInput).toBeInTheDocument();
        expect(nameInput).toHaveValue('');

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        expect(descriptionInput).toHaveValue('');

        const glyphInput = screen.queryByLabelText(/Apply Custom Glyph/);
        expect(glyphInput).not.toBeInTheDocument();

        // The Require Certification switch should not render on BHCE
        expect(screen.queryByText(/Require Certification/i)).not.toBeInTheDocument();

        // The delete button should not render when creating a new selector because it doesn't exist yet
        expect(screen.queryByRole('button', { name: /Delete/ })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Define Rule/ })).toBeInTheDocument();
        expect(screen.queryByText(/Enable Analysis/i)).not.toBeInTheDocument();
    });

    it('renders the form for editing an existing zone', async () => {
        // This url has the zone id of 1 in the path
        // and so this zone's data is filled into the form for the user to edit

        vi.mocked(useParams).mockReturnValue({ zoneId: '1', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: editHighestPrivilegeZonePath } as Location);

        render(
            <Routes>
                <Route path={editHighestPrivilegeZonePath} element={<TagForm />} />
            </Routes>,
            { route: editHighestPrivilegeZonePath }
        );

        await waitFor(() => {
            expect(screen.getByText('Edit Zone Details')).toBeInTheDocument();
        });

        const nameInput = await screen.findByLabelText('Zone Information');
        expect(nameInput).toBeInTheDocument();
        await waitFor(() => {
            expect(nameInput).toHaveValue('Tier Zero');
        });

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        await waitFor(() => {
            expect(descriptionInput).toHaveValue('Tier Zero Description');
        });

        // The Require Certification switch should not render on BHCE
        expect(screen.queryByText(/Require Certification/i)).not.toBeInTheDocument();

        // This form input is not available for the most privileged zone, aka Tier Zero
        expect(screen.queryByLabelText(/Apply Custom Glyph/)).not.toBeInTheDocument();

        // The delete button should not render when editing T0
        expect(screen.queryByRole('button', { name: /Delete/ })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Save/ })).toBeInTheDocument();
    });

    it('renders the form for editing an existing label', async () => {
        // This url has the label id of 2 in the path
        // and so this label's data is filled into the form for the user to edit

        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: '2' });
        vi.mocked(useLocation).mockReturnValue({ pathname: editExistingLabelPath } as Location);

        render(
            <Routes>
                <Route path={editExistingLabelPath} element={<TagForm />} />
            </Routes>,
            { route: editExistingLabelPath }
        );

        await waitFor(() => {
            expect(screen.getByText('Edit Label Details')).toBeInTheDocument();
        });

        const nameInput = await screen.findByLabelText('Label Information');
        expect(nameInput).toBeInTheDocument();
        await waitFor(() => {
            expect(nameInput).toHaveValue('Owned');
        });

        const descriptionInput = screen.getByLabelText('Description');
        expect(descriptionInput).toBeInTheDocument();
        await waitFor(() => {
            expect(descriptionInput).toHaveValue('Owned Description');
        });

        // The Require Certification switch should not render on BHCE
        expect(screen.queryByText(/Require Certification/i)).not.toBeInTheDocument();

        const glyphInput = screen.queryByLabelText(/Apply Custom Glyph/);
        expect(glyphInput).not.toBeInTheDocument();

        // The delete button should not render when editing Owned
        await waitFor(() => {
            expect(screen.queryByRole('button', { name: /Delete/ })).not.toBeInTheDocument();
        });
        expect(screen.getByRole('button', { name: /Cancel/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Save/ })).toBeInTheDocument();
        expect(screen.queryByTestId('privilege-zones_save_tag-form_analysis-enabled-switch')).not.toBeInTheDocument();
    });

    it('renders the glyph form input when the editing a zone that is not the highest privileged', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '4', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: editOtherZonePath } as Location);

        render(
            <Routes>
                <Route path={editOtherZonePath} element={<TagForm />} />
            </Routes>,
            { route: editOtherZonePath }
        );

        await waitFor(() => {
            expect(screen.getByText('Edit Zone Details')).toBeInTheDocument();
        });

        expect(await screen.findByLabelText(/Apply Custom Glyph/)).toBeInTheDocument();
        expect(await screen.findAllByText('lightbulb')).toHaveLength(2);
    });

    it('does not render the analysis toggle when multi tier analysis enabled is false', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '2', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: editHighestPrivilegeZonePath } as Location);

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
                <Route path={editHighestPrivilegeZonePath} element={<TagForm />} />
            </Routes>,
            { route: editHighestPrivilegeZonePath }
        );

        await waitFor(() => {
            expect(screen.getByText('Edit Zone Details')).toBeInTheDocument();
        });
        expect(screen.queryByText(/Enable Analysis/i)).not.toBeInTheDocument();
    });

    it('renders the analysis toggle when multi tier analysis enabled is true and when editing an existing zone', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '2', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: editHighestPrivilegeZonePath } as Location);

        render(
            <Routes>
                <Route path={editHighestPrivilegeZonePath} element={<TagForm />} />
            </Routes>,
            { route: editHighestPrivilegeZonePath }
        );

        await waitFor(() => {
            expect(screen.getByText('Edit Zone Details')).toBeInTheDocument();
        });
        expect(await screen.findByText(/Enable Analysis/i)).toBeInTheDocument();
    });

    test('clicking cancel on the form takes the user back to the page the user was on previously', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: createNewLabelPath } as Location);
        render(<TagForm />, { route: createNewLabelPath });

        await user.click(await screen.findByTestId('privilege-zones_save_tag-form_cancel-button'));

        expect(mockNavigate).toHaveBeenCalledWith(-1);
    });

    test('a name value is required to submit the form', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: createNewLabelPath } as Location);
        render(<TagForm />, { route: createNewLabelPath });

        await user.click(await screen.findByRole('button', { name: /Define Rule/ }));

        await waitFor(() => {
            expect(screen.getByText('Please provide a name for the Label')).toBeInTheDocument();
        });
    });

    it('validates that the name input is under the max length', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: createNewLabelPath } as Location);
        render(<TagForm />, { route: createNewLabelPath });

        const nameInput = await screen.findByLabelText('Label Information');

        await user.click(nameInput);
        await user.paste('f'.repeat(251));

        await user.click(await screen.findByRole('button', { name: /Define Rule/ }));

        await waitFor(() => {
            expect(
                screen.getByText(`Name cannot exceed 250 characters. Please provide a shorter name`)
            ).toBeInTheDocument();
        });
    });

    test('filling in the name value allows updating the selector and navigates back to the details page', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: createNewZonePath } as Location);
        render(
            <Routes>
                <Route path={'/'} element={<TagForm />} />
                <Route path={createNewZonePath} element={<TagForm />} />
            </Routes>,
            { route: createNewZonePath }
        );

        const nameInput = await screen.findByLabelText('Zone Information');

        await user.click(nameInput);
        await user.paste('foo');

        await user.click(await screen.findByRole('button', { name: /Define Rule/ }));

        expect(screen.queryByText('Please provide a name for the zone')).not.toBeInTheDocument();
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

        const nameInput = await screen.findByLabelText('Zone Information');

        await user.click(nameInput);
        await user.paste('foo');

        await user.click(await screen.findByRole('button', { name: /Define Rule/ }));

        expect(screen.queryByText('Please provide a name for the zone')).not.toBeInTheDocument();

        await waitFor(() => {
            expect(mockNavigate).toBeCalled();
        });
    });

    it('handles creating a new zone', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: createNewZonePath } as Location);

        render(
            <Routes>
                <Route path={'/'} element={<TagForm />} />
                <Route path={createNewZonePath} element={<TagForm />} />
            </Routes>,
            { route: createNewZonePath }
        );

        const nameInput = await screen.findByLabelText('Zone Information');

        await user.click(nameInput);
        await user.paste('foo');

        await user.click(await screen.findByRole('button', { name: /Define Rule/ }));

        await waitFor(() => {
            expect(mockNavigate).toBeCalled();
        });
    });

    it('handles creating a new label', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: createNewLabelPath } as Location);

        render(
            <Routes>
                <Route path={'/'} element={<TagForm />} />
                <Route path={createNewLabelPath} element={<TagForm />} />
            </Routes>,
            { route: createNewLabelPath }
        );

        const nameInput = await screen.findByLabelText('Label Information');

        await user.click(nameInput);
        await user.paste('foo');

        await user.click(await screen.findByRole('button', { name: /Define Rule/ }));

        await waitFor(() => {
            expect(mockNavigate).toBeCalled();
        });
    });

    it('only sends dirty fields in the request when updating', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '1', labelId: undefined });
        vi.mocked(useLocation).mockReturnValue({ pathname: editHighestPrivilegeZonePath } as Location);
        const updateAGTSpy = vi.spyOn(apiClient, 'updateAssetGroupTag');

        const queryClient = setUpQueryClient([{ key: privilegeZonesKeys.tagDetail('1'), data: testTierZero }]);

        render(
            <Routes>
                <Route path={'/'} element={<TagForm />} />
                <Route path={editHighestPrivilegeZonePath} element={<TagForm />} />
            </Routes>,
            { route: editHighestPrivilegeZonePath, queryClient }
        );

        const nameInput = await screen.findByLabelText('Zone Information');
        expect(nameInput).toBeInTheDocument();
        expect(nameInput).toHaveValue('Tier Zero');

        const descriptionInput = await screen.findByLabelText('Description');

        await user.click(descriptionInput);
        await user.clear(descriptionInput);
        await user.paste('updated field');

        await user.click(await screen.findByRole('button', { name: /Save Edits/ }));

        await waitFor(() => {
            expect(updateAGTSpy).toBeCalledWith('1', { description: 'updated field' }, undefined);
        });
    });

    it('disables the confirm button until user types required text', async () => {
        vi.mocked(useParams).mockReturnValue({ zoneId: '', labelId: '3' });
        vi.mocked(useLocation).mockReturnValue({ pathname: deletionTestsPath } as Location);

        render(<TagForm />);

        await screen.findByLabelText('Label Information');

        const deleteButton = screen.getByTestId('privilege-zones_save_tag-form_delete-button');

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
        vi.mocked(useLocation).mockReturnValue({ pathname: deletionTestsPath } as Location);

        render(<TagForm />);

        await screen.findByLabelText('Label Information');

        const deleteButton = screen.getByTestId('privilege-zones_save_tag-form_delete-button');

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
        vi.mocked(useLocation).mockReturnValue({ pathname: deletionTestsPath } as Location);
        server.use(
            rest.delete('/api/v2/asset-group-tags/:tagId', async (_, res, ctx) => {
                return res(ctx.status(200));
            })
        );

        render(
            <Routes>
                <Route path={deletionTestsPath} element={<TagForm />} />
            </Routes>,
            { route: deletionTestsPath }
        );

        await screen.findByLabelText('Label Information');

        const deleteButton = screen.getByRole('button', { name: /Delete Label/i });
        await act(async () => {
            fireEvent.click(deleteButton);
        });

        await waitFor(() => {
            expect(screen.getByRole('dialog')).toBeInTheDocument();
        });

        const confirmButton = screen.getByRole('button', { name: /confirm/i });
        expect(confirmButton).toBeDisabled();

        const textField = screen.getByTestId('confirmation-dialog_challenge-text');

        await act(async () => {
            await user.type(textField, 'Delete this label');
        });

        await waitFor(() => {
            expect(confirmButton).not.toBeDisabled();
        });

        await user.click(confirmButton);

        await waitFor(() => {
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        });
    });
});
