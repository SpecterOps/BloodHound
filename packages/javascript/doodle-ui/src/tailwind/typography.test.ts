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
import { PluginAPI } from 'tailwindcss/types/config';
import plugin from './plugin';
import preset from './preset';

describe('typography theme', () => {
    it('defines the approved font families and fallback stacks', () => {
        expect(preset.theme.fontFamily).toEqual({
            sans: ['Figtree', '"Segoe UI"', 'Helvetica', 'Arial', 'sans-serif'],
            heading: ['"Nunito Sans"', '"Avenir Next"', '"Segoe UI"', 'Helvetica', 'Arial', 'sans-serif'],
        });
    });

    it('exposes the muted text utility without removing the legacy text-light utility', () => {
        expect(preset.theme.extend.colors).toMatchObject({
            'text-light': 'var(--text-light)',
            'text-muted': 'var(--text-muted)',
        });
    });

    it('backs the light-mode muted token with the existing semantic value', () => {
        const addBase = vi.fn();

        plugin({
            addBase,
            addUtilities: vi.fn(),
        } as unknown as PluginAPI);

        const baseStyles = addBase.mock.calls[0][0];
        expect(baseStyles[' :root']['--text-muted']).toBe('var(--text-light)');
        expect(baseStyles[' :root']['--text-light']).toBe('#505050');
        expect(baseStyles['.dark']['--text-light']).toBe('#CDCDCD');
        expect(baseStyles['.dark']).not.toHaveProperty('--text-muted');
    });
});
