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

import { escapeCypherString } from './cypher';

describe('escapeCypherString', () => {
    it('wraps a plain value in single quotes', () => {
        expect(escapeCypherString('S-1-5-21-1234')).toBe("'S-1-5-21-1234'");
    });

    it('escapes a single quote', () => {
        expect(escapeCypherString("a'b")).toBe("'a\\'b'");
    });

    it('escapes multiple single quotes', () => {
        expect(escapeCypherString("'''")).toBe("'\\'\\'\\''");
    });

    it('escapes a backslash', () => {
        expect(escapeCypherString('a\\b')).toBe("'a\\\\b'");
    });

    it('escapes backslashes before quotes so they do not combine', () => {
        // Input: a\'b  (backslash + single quote)
        // Backslash must be escaped first to \\, then the quote to \'.
        // Result must be 'a\\\'b' so the parser reads literal `\` then literal `'`,
        // not an escaped quote that swallows the closing delimiter.
        expect(escapeCypherString("a\\'b")).toBe("'a\\\\\\'b'");
    });

    it('leaves double quotes untouched', () => {
        expect(escapeCypherString('a"b')).toBe("'a\"b'");
    });

    it('handles an empty string', () => {
        expect(escapeCypherString('')).toBe("''");
    });

    it('handles only escapable characters', () => {
        expect(escapeCypherString("\\'")).toBe("'\\\\\\''");
    });

    it('preserves unicode characters', () => {
        expect(escapeCypherString('ézoë')).toBe("'ézoë'");
    });
});
