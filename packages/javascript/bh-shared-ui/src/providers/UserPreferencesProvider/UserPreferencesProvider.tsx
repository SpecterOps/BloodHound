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
import { createContext, ReactNode, useCallback, useContext, useEffect, useRef, useState } from 'react';
import { useQuery, useQueryClient } from 'react-query';
import { useEnvironmentParams } from '../../hooks/useEnvironmentParams';
import { apiClient } from '../../utils/api';
import { useNotifications } from '../NotificationProvider';
import { decodeUserPreferences, defaultUserPreferences, encodeUserPreferences, UserPreferences } from './preferences';

export type PreferenceUpdate = Partial<Pick<UserPreferences, 'darkMode' | 'rememberDomain' | 'preferredDomainId'>>;

interface PreferencesContextValue {
    preferences: UserPreferences;
    isLoading: boolean;
    isReady: boolean;
    isSaving: boolean;
    loadError: boolean;
    saveError: boolean;
    updatePreferences: (update: PreferenceUpdate) => Promise<void>;
    retry: () => void;
}

const UserPreferencesContext = createContext<PreferencesContextValue>({
    preferences: defaultUserPreferences,
    isLoading: false,
    isReady: false,
    isSaving: false,
    loadError: false,
    saveError: false,
    updatePreferences: async () => {},
    retry: () => {},
});

export const useUserPreferences = () => useContext(UserPreferencesContext);

// Mount a separate provider for each authenticated user. Writes are serialized
// and merge against the most recently saved document, never a stale render.
export const UserPreferencesProvider = ({
    userId,
    onDarkModeChange,
    children,
}: {
    userId?: string;
    onDarkModeChange: (enabled: boolean) => void;
    children: ReactNode;
}) => {
    const queryClient = useQueryClient();
    const { addNotification } = useNotifications();
    const { environmentId } = useEnvironmentParams();
    const lastEnvironment = useRef<string | null>(null);
    const lifetime = useRef(new AbortController());
    const writes = useRef<Promise<void>>(Promise.resolve());
    const [pendingWrites, setPendingWrites] = useState(0);
    const [saveError, setSaveError] = useState(false);
    const query = useQuery(
        ['user-preferences', userId],
        ({ signal }) =>
            apiClient.getUserPreferences(userId!, { signal }).then((response) => decodeUserPreferences(response.data)),
        {
            enabled: !!userId,
            retry: false,
            cacheTime: 0,
            staleTime: Infinity,
            refetchOnWindowFocus: false,
        }
    );
    const preferences = query.data ?? defaultUserPreferences;

    useEffect(() => {
        const controller = new AbortController();
        lifetime.current = controller;
        return () => controller.abort();
    }, []);

    useEffect(() => {
        if (environmentId) lastEnvironment.current = environmentId;
    }, [environmentId]);

    useEffect(() => {
        onDarkModeChange(preferences.darkMode);
    }, [preferences.darkMode, onDarkModeChange]);

    const updatePreferences = useCallback(
        (patch: PreferenceUpdate): Promise<void> => {
            const controller = lifetime.current;
            const queryKey = ['user-preferences', userId];
            if (!userId || !queryClient.getQueryData<UserPreferences>(queryKey) || controller.signal.aborted) {
                return Promise.reject(new Error('User preferences have not loaded.'));
            }
            const update = { ...patch };
            if (patch.rememberDomain !== undefined) {
                update.preferredDomainId = patch.rememberDomain ? lastEnvironment.current : null;
            }
            setPendingWrites((count) => count + 1);
            setSaveError(false);
            const write = writes.current
                .catch(() => {})
                .then(async () => {
                    if (controller.signal.aborted) return;
                    const current = queryClient.getQueryData<UserPreferences>(queryKey);
                    if (!current) throw new Error('User preferences have not loaded.');
                    const next = { ...current, ...update };
                    await apiClient.upsertUserPreferences(userId, encodeUserPreferences(next), {
                        signal: controller.signal,
                    });
                    if (!controller.signal.aborted) queryClient.setQueryData(queryKey, next);
                })
                .catch((error) => {
                    if (!controller.signal.aborted) {
                        setSaveError(true);
                        addNotification(
                            'Unable to save user preferences. Please try again.',
                            'user-preferences-save-error',
                            { variant: 'error' }
                        );
                    }
                    throw error;
                })
                .finally(() => {
                    if (!controller.signal.aborted) setPendingWrites((count) => count - 1);
                });
            writes.current = write;
            return write;
        },
        [userId, queryClient, addNotification]
    );

    useEffect(() => {
        if (
            !query.isSuccess ||
            !preferences.rememberDomain ||
            !environmentId ||
            preferences.preferredDomainId === environmentId
        )
            return;
        const timer = window.setTimeout(() => {
            void updatePreferences({ preferredDomainId: environmentId }).catch(() => {});
        }, 300);
        return () => window.clearTimeout(timer);
    }, [query.isSuccess, preferences.rememberDomain, preferences.preferredDomainId, environmentId, updatePreferences]);

    return (
        <UserPreferencesContext.Provider
            value={{
                preferences,
                isLoading: !!userId && query.isLoading,
                isReady: !!userId && query.isSuccess,
                isSaving: pendingWrites > 0,
                loadError: query.isError,
                saveError,
                updatePreferences,
                retry: () => {
                    void query.refetch();
                },
            }}>
            {children}
        </UserPreferencesContext.Provider>
    );
};
