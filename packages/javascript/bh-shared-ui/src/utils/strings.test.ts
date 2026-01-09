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

import { isDomainLike } from './strings';

describe('isDomainLike', () => {
    it('returns true if the string looks like a domain', () => {
        expect(isDomainLike('example.com')).toBe(true);
        expect(isDomainLike('mail.google.com')).toBe(true);
        expect(isDomainLike('api.dev.example.com')).toBe(true);
        expect(isDomainLike('a.b.c.d.example.com')).toBe(true);
        expect(isDomainLike('123.456')).toBe(true);
        expect(isDomainLike('my-test_domain.co.uk')).toBe(true);
        expect(isDomainLike('test@domain.com')).toBe(true);
        expect(isDomainLike('user:pass@domain.com')).toBe(true);
        expect(isDomainLike('domain.com/path')).toBe(true);
        expect(isDomainLike('münchen.de')).toBe(true);
        expect(isDomainLike('日本.jp')).toBe(true);
        expect(isDomainLike('тест.рф')).toBe(true);
        expect(isDomainLike('Example.COM')).toBe(true);
        expect(isDomainLike('TEST.ORG')).toBe(true);
        expect(isDomainLike('MixedCase.Domain')).toBe(true);
    });

    it('returns false for strings without a period', () => {
        expect(isDomainLike('example')).toBe(false);
        expect(isDomainLike('localhost')).toBe(false);
        expect(isDomainLike('nodomain')).toBe(false);
    });

    it('returns false for strings with only whitespace around period', () => {
        expect(isDomainLike(' . ')).toBe(false);
        expect(isDomainLike('  .  ')).toBe(false);
    });

    it('returns false for strings with period at the start', () => {
        expect(isDomainLike('.example')).toBe(false);
        expect(isDomainLike('.com')).toBe(false);
    });

    it('returns false for strings with period at the end', () => {
        expect(isDomainLike('example.')).toBe(false);
        expect(isDomainLike('domain.')).toBe(false);
    });

    it('returns false for empty strings', () => {
        expect(isDomainLike('')).toBe(false);
    });

    it('returns false for strings with only a period', () => {
        expect(isDomainLike('.')).toBe(false);
    });

    it('returns false for strings with only whitespace', () => {
        expect(isDomainLike('   ')).toBe(false);
        expect(isDomainLike('\t')).toBe(false);
        expect(isDomainLike('\n')).toBe(false);
    });
});
