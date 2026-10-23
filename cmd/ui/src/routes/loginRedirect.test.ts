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

import {
    LOGIN_REDIRECT_STORAGE_KEY,
    clearLoginRedirect,
    getLoginDestination,
    saveLoginRedirect,
} from './loginRedirect';

const destination = '/explore?searchType=pathfinding#graph';

describe('login destinations', () => {
    beforeEach(() => sessionStorage.clear());
    afterEach(() => {
        vi.restoreAllMocks();
        sessionStorage.clear();
    });

    it('prefers the router destination over a saved destination without consuming storage', () => {
        sessionStorage.setItem(LOGIN_REDIRECT_STORAGE_KEY, destination);

        expect(
            getLoginDestination({ from: { pathname: '/my-profile', search: '?tab=details', hash: '#settings' } })
        ).toBe('/my-profile?tab=details#settings');
        expect(sessionStorage.getItem(LOGIN_REDIRECT_STORAGE_KEY)).toBe(destination);
    });

    it('restores a saved destination when router state is absent', () => {
        sessionStorage.setItem(LOGIN_REDIRECT_STORAGE_KEY, destination);

        expect(getLoginDestination(null)).toBe(destination);
        expect(sessionStorage.getItem(LOGIN_REDIRECT_STORAGE_KEY)).toBe(destination);
    });

    it.each(['/', '/login', '/expired-password', '/user-disabled', '//example.com', '/\\example.com'])(
        'uses the saved destination when the router destination is unsafe or would loop: %s',
        (pathname) => {
            sessionStorage.setItem(LOGIN_REDIRECT_STORAGE_KEY, destination);

            expect(getLoginDestination({ from: { pathname, search: '', hash: '' } })).toBe(destination);
        }
    );

    it.each(['/', '/?query=value', '//example.com', '/\\example.com', '/login', '/expired-password', '/user-disabled'])(
        'falls back when the saved destination is unsafe or would loop: %s',
        (savedDestination) => {
            sessionStorage.setItem(LOGIN_REDIRECT_STORAGE_KEY, savedDestination);

            expect(getLoginDestination(null, '/fallback')).toBe('/fallback');
        }
    );

    it('defaults to Explore when no destination exists', () => {
        expect(getLoginDestination(null)).toBe('/explore');
    });

    it('uses the fallback when reading storage fails', () => {
        vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('Storage blocked');
        });

        expect(getLoginDestination(null, '/fallback')).toBe('/fallback');
    });

    it('allows SSO to continue when saving the destination fails', () => {
        vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('Storage blocked');
        });

        expect(() => saveLoginRedirect(destination)).not.toThrow();
    });

    it('allows navigation to continue when clearing storage fails', () => {
        vi.spyOn(Storage.prototype, 'removeItem').mockImplementation(() => {
            throw new Error('Storage blocked');
        });

        expect(() => clearLoginRedirect()).not.toThrow();
    });
});
