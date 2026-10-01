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
import { decodeUserPreferences, defaultUserPreferences, encodeUserPreferences } from './preferences';

describe('user preference encoding', () => {
    it('treats empty storage as defaults', () => {
        expect(decodeUserPreferences(new ArrayBuffer(0))).toEqual(defaultUserPreferences);
    });
    it('round-trips UTF-8 and preserves unknown settings', () => {
        const preferences = { ...defaultUserPreferences, darkMode: true, futureSetting: '日本語' };
        expect(decodeUserPreferences(encodeUserPreferences(preferences))).toEqual(preferences);
    });
    it.each(['null', '[]', '{', '{"version":2}', JSON.stringify({ ...defaultUserPreferences, darkMode: 'true' })])(
        'rejects invalid or unsupported data: %s',
        (value) => {
            expect(() => decodeUserPreferences(new TextEncoder().encode(value).buffer)).toThrow();
        }
    );
    it('enforces the byte limit for multibyte strings', () => {
        expect(() => encodeUserPreferences({ ...defaultUserPreferences, large: '界'.repeat(23000) })).toThrow(
            'storage limit'
        );
    });
});
