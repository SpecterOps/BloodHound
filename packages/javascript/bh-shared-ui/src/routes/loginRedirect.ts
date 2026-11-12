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

import { Path, createPath } from 'react-router-dom';

export const LOGIN_REDIRECT_STORAGE_KEY = 'loginRedirect';

export type LoginRedirectState = {
    from?: Path;
};

export type LoginRedirectOptions = {
    fallback: string;
    excludedPaths: readonly string[];
};

/** Preserve the original URL when authentication adds another intermediate route. */
export const getLoginRedirectState = (location: Path & { state: LoginRedirectState | null }): LoginRedirectState => ({
    from: location.state?.from ?? {
        pathname: location.pathname,
        search: location.search,
        hash: location.hash,
    },
});

const isValidDestination = (
    destination: string | undefined,
    origin: string,
    excludedPaths: readonly string[]
): destination is string => {
    if (!destination || !destination.startsWith('/') || destination.startsWith('//') || destination.includes('\\'))
        return false;

    const url = new URL(destination, origin);
    return url.origin === origin && !excludedPaths.includes(url.pathname);
};

export const saveLoginRedirect = (destination: string): void => {
    try {
        sessionStorage.setItem(LOGIN_REDIRECT_STORAGE_KEY, destination);
    } catch {
        // Authentication must still work when browser storage is unavailable.
    }
};

const getSavedDestination = (): string | undefined => {
    try {
        return sessionStorage.getItem(LOGIN_REDIRECT_STORAGE_KEY) ?? undefined;
    } catch {
        return undefined;
    }
};

export const clearLoginRedirect = (): void => {
    try {
        sessionStorage.removeItem(LOGIN_REDIRECT_STORAGE_KEY);
    } catch {
        // Authentication must still work when browser storage is unavailable.
    }
};

type ResolveLoginDestinationInput = {
    state: LoginRedirectState | null;
    savedDestination: string | undefined;
    origin: string;
    options: LoginRedirectOptions;
};

export const resolveLoginDestination = ({
    state,
    savedDestination,
    origin,
    options: { fallback, excludedPaths },
}: ResolveLoginDestinationInput): string => {
    const routeDestination = state?.from && createPath(state.from);
    if (isValidDestination(routeDestination, origin, excludedPaths)) return routeDestination;

    if (isValidDestination(savedDestination, origin, excludedPaths)) return savedDestination;

    return fallback;
};

/** Resolves without consuming storage; clear it only after all authentication requirements are met. */
export const getLoginDestination = (state: LoginRedirectState | null, options: LoginRedirectOptions): string =>
    resolveLoginDestination({
        state,
        savedDestination: getSavedDestination(),
        origin: window.location.origin,
        options,
    });
