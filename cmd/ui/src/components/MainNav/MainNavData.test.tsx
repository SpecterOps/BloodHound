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

import { render, renderHook, screen } from 'src/test-utils';
import { useMainNavSecondaryListData } from './MainNavData';

vi.mock('bh-shared-ui', async (importOriginal) => {
    const actual = await importOriginal<typeof import('bh-shared-ui')>();

    return {
        ...actual,
        useKeybindings: vi.fn(),
        useSubNavRoutes: () => ({ routes: [] }),
    };
});

describe('MainNavData', () => {
    it('places Marketplace above Profile with the four-square icon', () => {
        const { result } = renderHook(() => useMainNavSecondaryListData());
        const labels = result.current.map(({ label }) => label);
        const marketplaceIndex = labels.indexOf('Marketplace');
        const profileIndex = labels.indexOf('Profile');

        expect(marketplaceIndex).toBeLessThan(profileIndex);
        expect(result.current[marketplaceIndex].route).toBe('/marketplace');

        const { container } = render(<>{result.current[marketplaceIndex].icon}</>);
        expect(screen.getByTestId('marketplace-icon')).toBeInTheDocument();
        expect(container.querySelectorAll('rect')).toHaveLength(4);
    });
});
