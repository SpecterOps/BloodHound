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

import React from 'react';
import { renderHook } from '../test-utils';
import { SubNavItem, SubNavSection } from '../types';
import { Permission } from '../utils/permissions';
import { useSubNavRoutes } from './useSubNavRoutes';

const mockUseFeatureFlags = vi.fn();
const mockUsePermissions = vi.fn();

vi.mock('./useFeatureFlags', async () => {
    const actual = await vi.importActual<typeof import('./useFeatureFlags')>('./useFeatureFlags');
    return {
        ...actual,
        useFeatureFlags: (...args: unknown[]) => mockUseFeatureFlags(...args),
    };
});

vi.mock('./usePermissions', () => ({
    usePermissions: () => mockUsePermissions(),
}));

const dummyComponent = React.lazy(() => Promise.resolve({ default: () => null }));

const makeItem = (overrides: Partial<SubNavItem> = {}): SubNavItem => ({
    label: 'item',
    path: '/item',
    component: dummyComponent,
    adminOnly: false,
    ...overrides,
});

const setFeatureFlags = ({ enabledKeys = [] as string[], isLoading = false } = {}) => {
    const data = enabledKeys.map((key) => ({ key, enabled: true }));
    mockUseFeatureFlags.mockReturnValue({ data, isLoading });
};

const setPermissions = ({ granted = [] as Permission[], isLoading = false } = {}) => {
    mockUsePermissions.mockReturnValue({
        isLoading,
        checkAllPermissions: (perms: Permission[]) => perms.every((p) => granted.includes(p)),
    });
};

const ADMIN_PERMS = [Permission.APP_READ_APPLICATION_CONFIGURATION, Permission.APP_WRITE_APPLICATION_CONFIGURATION];

const renderRoutes = (sections: SubNavSection[]) =>
    renderHook(() => useSubNavRoutes(sections, true)).result.current.routes;

describe('useSubNavRoutes', () => {
    beforeEach(() => {
        setFeatureFlags();
        setPermissions();
    });

    describe('featureFlag gate', () => {
        const sections: SubNavSection[] = [{ title: 's', items: [makeItem({ featureFlag: 'flag_a' })] }];

        it('hides items when the flag is disabled', () => {
            expect(renderRoutes(sections)).toEqual([]);
        });

        it('shows items when the flag is enabled', () => {
            setFeatureFlags({ enabledKeys: ['flag_a'] });
            expect(renderRoutes(sections)[0].items).toHaveLength(1);
        });

        it('hides items while feature flags are loading', () => {
            setFeatureFlags({ enabledKeys: ['flag_a'], isLoading: true });
            expect(renderRoutes(sections)).toEqual([]);
        });
    });

    describe('permissions gate', () => {
        const sections: SubNavSection[] = [
            { title: 's', items: [makeItem({ permissions: [Permission.ALERTS_READ] })] },
        ];

        it('hides items when the user is missing a required permission', () => {
            expect(renderRoutes(sections)).toEqual([]);
        });

        it('shows items when all required permissions are granted', () => {
            setPermissions({ granted: [Permission.ALERTS_READ] });
            expect(renderRoutes(sections)[0].items).toHaveLength(1);
        });

        it('hides items while permissions are loading', () => {
            setPermissions({ granted: [Permission.ALERTS_READ], isLoading: true });
            expect(renderRoutes(sections)).toEqual([]);
        });
    });

    describe('adminOnly gate', () => {
        const sections: SubNavSection[] = [{ title: 's', items: [makeItem({ adminOnly: true })] }];

        it('hides items when the user lacks admin permissions', () => {
            expect(renderRoutes(sections)).toEqual([]);
        });

        it('shows items when the user has admin permissions', () => {
            setPermissions({ granted: ADMIN_PERMS });
            expect(renderRoutes(sections)[0].items).toHaveLength(1);
        });

        it('hides items while permissions are loading', () => {
            setPermissions({ granted: ADMIN_PERMS, isLoading: true });
            expect(renderRoutes(sections)).toEqual([]);
        });
    });

    describe('combined gates', () => {
        const item = makeItem({ featureFlag: 'flag_a', permissions: [Permission.ALERTS_READ] });
        const sections: SubNavSection[] = [{ title: 's', items: [item] }];

        it('hides the item when only the feature flag passes', () => {
            setFeatureFlags({ enabledKeys: ['flag_a'] });
            expect(renderRoutes(sections)).toEqual([]);
        });

        it('hides the item when only the permission passes', () => {
            setPermissions({ granted: [Permission.ALERTS_READ] });
            expect(renderRoutes(sections)).toEqual([]);
        });

        it('shows the item only when both gates pass', () => {
            setFeatureFlags({ enabledKeys: ['flag_a'] });
            setPermissions({ granted: [Permission.ALERTS_READ] });
            expect(renderRoutes(sections)[0].items).toHaveLength(1);
        });
    });

    it('omits sections whose items are all filtered out', () => {
        const sections: SubNavSection[] = [
            { title: 'visible', items: [makeItem()] },
            { title: 'hidden', items: [makeItem({ featureFlag: 'flag_a' })] },
        ];
        const routes = renderRoutes(sections);
        expect(routes).toHaveLength(1);
        expect(routes[0].title).toBe('visible');
    });
});
