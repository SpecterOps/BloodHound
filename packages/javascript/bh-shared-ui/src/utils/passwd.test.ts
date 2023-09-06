// Copyright 2023 Specter Ops, Inc.
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

import { testPassword } from './passwd';

describe('testPassword', () => {
    it('succeeds for valid password', () => {
        expect(testPassword('Test123$!@#%')).toBe(true);
    });

    it('succeeds for valid password with unicode characters', () => {
        expect(testPassword('Test1234567⏥')).toBe(true);
    });

    it('fails for password without lowercase letters', () => {
        expect(testPassword('TEST123$!@#%')).toBe(false);
    });

    it('fails for password without uppercase letters', () => {
        expect(testPassword('test123$!@#%')).toBe(false);
    });

    it('fails for password without digits', () => {
        expect(testPassword('Testabc$!@#%')).toBe(false);
    });

    it('fails for password without special characters', () => {
        expect(testPassword('Test12345678')).toBe(false);
    });

    it('fails for password shorter than 12 characters', () => {
        expect(testPassword('Test123$!')).toBe(false);
    });

    it('fails for empty password', () => {
        expect(testPassword('')).toBe(false);
    });
});
