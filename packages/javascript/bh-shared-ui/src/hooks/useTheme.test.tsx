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
import { act, cleanup, renderHook, waitFor } from '@testing-library/react';
import { darkTheme, lightTheme, useTheme } from './useTheme';

describe('useTheme', () => {
    let originalClassName: string;
    let originalStyle: string | null;

    beforeEach(() => {
        originalClassName = document.documentElement.className;
        originalStyle = document.documentElement.getAttribute('style');
    });

    afterEach(() => {
        cleanup();
        document.documentElement.className = originalClassName;
        if (originalStyle === null) {
            document.documentElement.removeAttribute('style');
        } else {
            document.documentElement.setAttribute('style', originalStyle);
        }
    });

    it('includes light and dark defaults', () => {
        expect(lightTheme.primaryVariant).toBe('#0D0A30');
        expect(lightTheme).not.toHaveProperty('isDarkMode');
        expect(darkTheme.primaryVariant).toBe('#8D8BF8');
        expect(darkTheme).not.toHaveProperty('isDarkMode');
    });

    it.each([
        ['light', false],
        ['dark', true],
        ['', false],
        ['dark other-class', true],
    ])('reads dark-mode status from the initial root class "%s"', (className, isDarkMode) => {
        document.documentElement.className = className;

        const { result } = renderHook(() => useTheme());

        expect(result.current.isDarkMode).toBe(isDarkMode);
    });

    it.each(['light', 'dark'])('reads the computed token on mount in %s mode', (mode) => {
        document.documentElement.className = mode;
        document.documentElement.style.setProperty('--primary-variant', '#123456');

        const { result } = renderHook(() => useTheme());

        expect(result.current.primaryVariant).toBe('#123456');
    });

    it('updates the token and dark-mode status when switching themes', async () => {
        document.documentElement.className = 'light';
        document.documentElement.style.setProperty('--primary-variant', lightTheme.primaryVariant);

        const { result } = renderHook(() => useTheme());

        expect(result.current.primaryVariant).toBe(lightTheme.primaryVariant);
        expect(result.current.isDarkMode).toBe(false);

        act(() => {
            document.documentElement.style.setProperty('--primary-variant', darkTheme.primaryVariant);
            document.documentElement.className = 'dark';
        });

        await waitFor(() => {
            expect(result.current.primaryVariant).toBe(darkTheme.primaryVariant);
            expect(result.current.isDarkMode).toBe(true);
        });

        act(() => {
            document.documentElement.style.setProperty('--primary-variant', lightTheme.primaryVariant);
            document.documentElement.className = 'light';
        });

        await waitFor(() => {
            expect(result.current.primaryVariant).toBe(lightTheme.primaryVariant);
            expect(result.current.isDarkMode).toBe(false);
        });
    });
});
