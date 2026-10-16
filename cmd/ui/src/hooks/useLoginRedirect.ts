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

import { useEffect, useMemo } from 'react';
import { useLocation } from 'react-router-dom';
import { ROUTE_EXPIRED_PASSWORD, ROUTE_HOME, ROUTE_LOGIN, ROUTE_USER_DISABLED } from 'src/routes/constants';

export const LOGIN_REDIRECT_STORAGE_KEY = 'loginRedirect';

type RedirectLocationState = {
    from?: {
        pathname: string;
        search?: string;
        hash?: string;
    };
};

const isValidDestination = (destination: string) => {
    if (!destination.startsWith('/') || destination.startsWith('//') || destination.includes('\\')) return false;

    const url = new URL(destination, window.location.origin);
    return (
        url.origin === window.location.origin &&
        ![ROUTE_HOME, ROUTE_LOGIN, ROUTE_EXPIRED_PASSWORD, ROUTE_USER_DISABLED].includes(url.pathname)
    );
};

export const saveLoginRedirect = (destination: string) => {
    try {
        sessionStorage.setItem(LOGIN_REDIRECT_STORAGE_KEY, destination);
    } catch {
        // Authentication must still work when browser storage is unavailable.
    }
};

const getSavedDestination = () => {
    try {
        const destination = sessionStorage.getItem(LOGIN_REDIRECT_STORAGE_KEY);
        return destination && isValidDestination(destination) ? destination : undefined;
    } catch {
        return undefined;
    }
};

/** Preserves an SSO destination across page loads and consumes it after authentication. */
export const useLoginRedirect = (readyToRedirect: boolean, fallback = ROUTE_HOME) => {
    const location = useLocation();
    const destination = useMemo(() => {
        const from = (location.state as RedirectLocationState | null)?.from;
        const routeDestination = from ? `${from.pathname}${from.search ?? ''}${from.hash ?? ''}` : undefined;
        // SSO returns to the home route, which can then redirect to password reset.
        if (routeDestination && from?.pathname !== ROUTE_HOME && isValidDestination(routeDestination)) {
            return routeDestination;
        }
        return (
            getSavedDestination() ??
            (routeDestination && isValidDestination(routeDestination) ? routeDestination : fallback)
        );
    }, [location.state, fallback]);

    useEffect(() => {
        if (readyToRedirect) {
            try {
                sessionStorage.removeItem(LOGIN_REDIRECT_STORAGE_KEY);
            } catch {
                // Authentication must still work when browser storage is unavailable.
            }
        }
    }, [readyToRedirect]);

    return destination;
};
