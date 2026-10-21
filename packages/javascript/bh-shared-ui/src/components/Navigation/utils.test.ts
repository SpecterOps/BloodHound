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

import { isRouteActive } from './utils';

describe('isRouteActive', () => {
    it.each([
        [undefined, '/'],
        [undefined, '/test'],
        ['', '/'],
        ['', '/test'],
    ])('returns false when route is %j (pathname %s)', (route, pathname) => {
        expect(isRouteActive(pathname, route)).toBe(false);
    });

    it.each([
        ['/', '/', true],
        ['/', '/test', false],
        ['/*', '/', true],
        ['/*', '/test', false],
        ['/test', '/test', true],
        ['/test', '/test/', true],
        ['/test', '/test/child', true],
        ['/test', '/test/child/grandchild', true],
        ['/test/', '/test', true],
        ['/test/', '/test/child', true],
        ['/test//', '/test', true],
        ['/test/*', '/test', true],
        ['/test/*', '/test/child', true],
        ['/test', '/test-other', false],
        ['/test', '/testing', false],
        ['/test', '/other/test', false],
        ['/test', '/', false],
        ['/test/*', '/test-other', false],
        ['/test/child', '/test', false],
    ])('route %s at pathname %s returns %s', (route, pathname, expected) => {
        expect(isRouteActive(pathname, route)).toBe(expected);
    });
});
