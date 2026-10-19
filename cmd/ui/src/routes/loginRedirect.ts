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
import { ROUTE_EXPIRED_PASSWORD, ROUTE_EXPLORE, ROUTE_HOME, ROUTE_LOGIN, ROUTE_USER_DISABLED } from './constants';

export const LOGIN_REDIRECT_STORAGE_KEY = 'loginRedirect';

export type LoginRedirectState = {
    from?: Path;
};

const isValidDestination = (destination: string | undefined): destination is string => {
    if (!destination || !destination.startsWith('/') || destination.startsWith('//') || destination.includes('\\'))
        return false;

    const url = new URL(destination, window.location.origin);
    return (
        url.origin === window.location.origin &&
        ![ROUTE_HOME, ROUTE_LOGIN, ROUTE_EXPIRED_PASSWORD, ROUTE_USER_DISABLED].includes(url.pathname)
    );
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

/** Resolves the original destination without consuming it. Only Home clears the saved destination. */
export const getLoginDestination = (state: LoginRedirectState | null, fallback = ROUTE_EXPLORE): string => {
    const routeDestination = state?.from && createPath(state.from);
    if (isValidDestination(routeDestination)) return routeDestination;

    const savedDestination = getSavedDestination();
    if (isValidDestination(savedDestination)) return savedDestination;

    return fallback;
};
