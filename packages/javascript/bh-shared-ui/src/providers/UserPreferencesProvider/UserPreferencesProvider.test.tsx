// Copyright 2026 Specter Ops, Inc.
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
import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AxiosResponse } from 'axios';
import { HelmetProvider } from 'react-helmet-async';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter, useNavigate } from 'react-router-dom';
import { useInitialEnvironment } from '../../hooks/useInitialEnvironment';
import { apiClient } from '../../utils/api';
import UserPreferences from '../../views/UserPreferences/UserPreferences';
import { decodeUserPreferences, defaultUserPreferences, encodeUserPreferences } from './preferences';
import { UserPreferencesProvider, useUserPreferences } from './UserPreferencesProvider';

const getPreferences = vi.spyOn(apiClient, 'getUserPreferences');
const putPreferences = vi.spyOn(apiClient, 'upsertUserPreferences');
const getEnvironments = vi.spyOn(apiClient, 'getAvailableEnvironments');
const response = (data: ArrayBuffer) => ({ data }) as AxiosResponse<ArrayBuffer>;
const onDarkModeChange = vi.fn();

const Probe = () => {
    const navigate = useNavigate();
    const { preferences, updatePreferences, isReady } = useUserPreferences();
    return (
        <>
            <output data-testid='document'>{JSON.stringify(preferences)}</output>
            <button onClick={() => navigate('/?environmentId=domain-b')}>Select domain B</button>
            <button
                disabled={!isReady}
                onClick={() => {
                    void updatePreferences({ darkMode: true }).catch(() => {});
                    void updatePreferences({ rememberDomain: true }).catch(() => {});
                }}>
                Queue changes
            </button>
        </>
    );
};

const InitialEnvironment = ({ disabled = false }: { disabled?: boolean }) => {
    const { data } = useInitialEnvironment({ orderBy: 'name', queryOptions: { enabled: !disabled } });
    return <output data-testid='initial-domain'>{data?.id}</output>;
};

const createClient = () =>
    new QueryClient({
        defaultOptions: { queries: { retry: false } },
    });
const Harness = ({
    userId = 'user-a',
    queryClient,
    initialEnvironment = false,
    disabled = false,
}: {
    userId?: string;
    queryClient: QueryClient;
    initialEnvironment?: boolean;
    disabled?: boolean;
}) => (
    <QueryClientProvider client={queryClient}>
        <MemoryRouter>
            <HelmetProvider>
                <UserPreferencesProvider key={userId} userId={userId} onDarkModeChange={onDarkModeChange}>
                    <UserPreferences />
                    <Probe />
                    {initialEnvironment && <InitialEnvironment disabled={disabled} />}
                </UserPreferencesProvider>
            </HelmetProvider>
        </MemoryRouter>
    </QueryClientProvider>
);

beforeEach(() => {
    getPreferences.mockReset().mockResolvedValue(response(new ArrayBuffer(0)));
    putPreferences.mockReset().mockResolvedValue({ status: 204 } as AxiosResponse<void>);
    getEnvironments.mockReset().mockResolvedValue({
        data: {
            data: [
                { id: 'domain-a', name: 'A', type: 'active-directory', collected: true, impactValue: 1 },
                { id: 'domain-b', name: 'B', type: 'active-directory', collected: true, impactValue: 2 },
            ],
        },
    } as Awaited<ReturnType<typeof apiClient.getAvailableEnvironments>>);
});

describe('User Preferenecs', () => {
    it('shows the requested table and MFA-style switches and saves the full binary document', async () => {
        const user = userEvent.setup();
        render(<Harness queryClient={createClient()} />);
        expect(screen.getByRole('heading', { name: 'User Preferenecs' })).toBeInTheDocument();
        expect(screen.getByRole('table', { name: 'User preferences' })).toBeInTheDocument();
        const darkMode = screen.getByRole('switch', { name: 'Dark Mode' });
        await waitFor(() => expect(darkMode).toBeEnabled());
        await user.click(darkMode);
        await waitFor(() => expect(darkMode).toBeChecked());
        expect(decodeUserPreferences(putPreferences.mock.calls[0][1])).toEqual({
            ...defaultUserPreferences,
            darkMode: true,
        });
        expect(putPreferences.mock.calls[0][0]).toBe('user-a');
        expect(onDarkModeChange).toHaveBeenLastCalledWith(true);
    });

    it('loads saved settings without writing defaults over them', async () => {
        getPreferences.mockResolvedValue(
            response(encodeUserPreferences({ ...defaultUserPreferences, darkMode: true, rememberDomain: true }))
        );
        render(<Harness queryClient={createClient()} />);
        await waitFor(() => expect(screen.getByRole('switch', { name: 'Dark Mode' })).toBeChecked());
        expect(screen.getByRole('switch', { name: 'Preferred Domain' })).toBeChecked();
        expect(putPreferences).not.toHaveBeenCalled();
    });

    it('keeps saved values on write failure and lets the user retry', async () => {
        const user = userEvent.setup();
        putPreferences.mockRejectedValueOnce(new Error('offline'));
        render(<Harness queryClient={createClient()} />);
        const toggle = screen.getByRole('switch', { name: 'Dark Mode' });
        await waitFor(() => expect(toggle).toBeEnabled());
        await user.click(toggle);
        await screen.findByText('Your change could not be saved. Please try again.');
        expect(toggle).not.toBeChecked();
        await user.click(toggle);
        await waitFor(() => expect(toggle).toBeChecked());
    });

    it('does not overwrite malformed or newer documents and can retry loading', async () => {
        const user = userEvent.setup();
        getPreferences.mockResolvedValueOnce(response(new TextEncoder().encode('{"version":2}').buffer));
        render(<Harness queryClient={createClient()} />);
        await screen.findByText('Unable to load your preferences. Your saved settings have not been changed.');
        expect(screen.getByRole('switch', { name: 'Dark Mode' })).toBeDisabled();
        expect(putPreferences).not.toHaveBeenCalled();
        await user.click(screen.getByRole('button', { name: 'Retry' }));
        await waitFor(() => expect(screen.getByRole('switch', { name: 'Dark Mode' })).toBeEnabled());
    });

    it('serializes writes and preserves both changes and unknown fields', async () => {
        const user = userEvent.setup();
        getPreferences.mockResolvedValue(
            response(encodeUserPreferences({ ...defaultUserPreferences, futureSetting: 'preserve me' }))
        );
        let completeFirst!: () => void;
        putPreferences.mockImplementationOnce(
            () =>
                new Promise((resolve) => {
                    completeFirst = () => resolve({ status: 204 } as AxiosResponse<void>);
                })
        );
        render(<Harness queryClient={createClient()} />);
        const queue = screen.getByRole('button', { name: 'Queue changes' });
        await waitFor(() => expect(queue).toBeEnabled());
        await user.click(queue);
        await waitFor(() => expect(putPreferences).toHaveBeenCalledTimes(1));
        await act(async () => completeFirst());
        await waitFor(() => expect(putPreferences).toHaveBeenCalledTimes(2));
        expect(decodeUserPreferences(putPreferences.mock.calls[1][1])).toMatchObject({
            darkMode: true,
            rememberDomain: true,
            futureSetting: 'preserve me',
        });
    });

    it('remembers domain changes only when enabled and clears the saved domain when disabled', async () => {
        const user = userEvent.setup();
        render(<Harness queryClient={createClient()} />);
        const toggle = screen.getByRole('switch', { name: 'Preferred Domain' });
        await waitFor(() => expect(toggle).toBeEnabled());
        await user.click(screen.getByRole('button', { name: 'Select domain B' }));
        expect(putPreferences).not.toHaveBeenCalled();
        await user.click(toggle);
        await waitFor(() => expect(toggle).toBeChecked());
        expect(decodeUserPreferences(putPreferences.mock.calls[0][1])).toMatchObject({
            rememberDomain: true,
            preferredDomainId: 'domain-b',
        });
        await user.click(toggle);
        await waitFor(() => expect(toggle).not.toBeChecked());
        expect(decodeUserPreferences(putPreferences.mock.calls[1][1])).toMatchObject({
            rememberDomain: false,
            preferredDomainId: null,
        });
    });

    it('persists later domain navigation when remembering is enabled', async () => {
        const user = userEvent.setup();
        getPreferences.mockResolvedValue(
            response(
                encodeUserPreferences({
                    ...defaultUserPreferences,
                    rememberDomain: true,
                    preferredDomainId: 'domain-a',
                })
            )
        );
        render(<Harness queryClient={createClient()} />);
        await waitFor(() => expect(screen.getByRole('switch', { name: 'Preferred Domain' })).toBeEnabled());
        await user.click(screen.getByRole('button', { name: 'Select domain B' }));
        await waitFor(() => expect(putPreferences).toHaveBeenCalledTimes(1));
        expect(decodeUserPreferences(putPreferences.mock.calls[0][1]).preferredDomainId).toBe('domain-b');
    });

    it.each(['domain-b', 'unavailable-domain'])(
        'uses a saved accessible domain or falls back safely: %s',
        async (savedDomain) => {
            getPreferences.mockResolvedValue(
                response(
                    encodeUserPreferences({
                        ...defaultUserPreferences,
                        rememberDomain: true,
                        preferredDomainId: savedDomain,
                    })
                )
            );
            render(<Harness queryClient={createClient()} initialEnvironment />);
            await waitFor(() =>
                expect(screen.getByTestId('initial-domain')).toHaveTextContent(
                    savedDomain === 'domain-b' ? 'domain-b' : 'domain-a'
                )
            );
        }
    );

    it('does not select an initial domain while preferences are loading or selection is disabled', async () => {
        getPreferences.mockReturnValue(new Promise(() => {}));
        const view = render(<Harness queryClient={createClient()} initialEnvironment />);
        expect(getEnvironments).not.toHaveBeenCalled();
        view.unmount();
        getPreferences.mockResolvedValue(
            response(
                encodeUserPreferences({
                    ...defaultUserPreferences,
                    rememberDomain: true,
                    preferredDomainId: 'domain-b',
                })
            )
        );
        render(<Harness queryClient={createClient()} initialEnvironment disabled />);
        await waitFor(() => expect(screen.getByRole('switch', { name: 'Dark Mode' })).toBeEnabled());
        expect(getEnvironments).not.toHaveBeenCalled();
    });

    it('aborts old writes and isolates another account from pending changes', async () => {
        const user = userEvent.setup();
        const queryClient = createClient();
        let completeFirst!: () => void;
        putPreferences.mockImplementationOnce(
            () =>
                new Promise((resolve) => {
                    completeFirst = () => resolve({ status: 204 } as AxiosResponse<void>);
                })
        );
        const view = render(<Harness queryClient={queryClient} />);
        await waitFor(() => expect(screen.getByRole('button', { name: 'Queue changes' })).toBeEnabled());
        await user.click(screen.getByRole('button', { name: 'Queue changes' }));
        await waitFor(() => expect(putPreferences).toHaveBeenCalledTimes(1));
        const signal = putPreferences.mock.calls[0][2]?.signal;
        view.rerender(<Harness queryClient={queryClient} userId='user-b' />);
        expect(signal?.aborted).toBe(true);
        await act(async () => completeFirst());
        await waitFor(() => expect(screen.getByRole('switch', { name: 'Dark Mode' })).toBeEnabled());
        expect(screen.getByRole('switch', { name: 'Dark Mode' })).not.toBeChecked();
        expect(putPreferences).toHaveBeenCalledTimes(1);
        expect(getPreferences).toHaveBeenLastCalledWith('user-b', expect.anything());
    });
});
