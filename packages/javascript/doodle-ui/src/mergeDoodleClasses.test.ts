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

import { cn, cnWithState } from './components/utils';
import { mergeDoodleClasses } from './mergeDoodleClasses';

describe('mergeDoodleClasses', () => {
    it('is used by the component class helpers', () => {
        expect(cn('text-body-2', { 'text-text-main': true })).toBe('text-body-2 text-text-main');
        expect(
            cnWithState<{ active: boolean }>(
                'text-body-2',
                ({ active }) => active && 'text-text-main'
            )({ active: true })
        ).toBe('text-body-2 text-text-main');
    });
    it.each(['headline-1', 'headline-2', 'headline-3', 'body', 'body-2', 'button', 'caption', 'eyeline'])(
        'keeps the %s font size independent of text color',
        (size) => {
            expect(mergeDoodleClasses(`text-${size}`, 'text-text-main')).toBe(`text-${size} text-text-main`);
            expect(mergeDoodleClasses('text-text-main', `text-${size}`)).toBe(`text-text-main text-${size}`);
        }
    );

    it('lets either core or design-system font sizes override the other', () => {
        expect(mergeDoodleClasses('text-base', 'text-body-2')).toBe('text-body-2');
        expect(mergeDoodleClasses('text-body-2', 'text-base')).toBe('text-base');
        expect(mergeDoodleClasses('text-body-2', 'text-caption')).toBe('text-caption');
    });

    it('preserves independent responsive and state variants', () => {
        expect(mergeDoodleClasses('text-body dark:text-caption', 'dark:text-text-main sm:text-body-2')).toBe(
            'text-body dark:text-caption dark:text-text-main sm:text-body-2'
        );
        expect(mergeDoodleClasses('sm:text-body', 'sm:text-caption')).toBe('sm:text-caption');
    });

    it('merges explicit line heights and font-size modifiers in order', () => {
        expect(mergeDoodleClasses('text-caption/4', 'text-text-muted')).toBe('text-caption/4 text-text-muted');
        expect(mergeDoodleClasses('leading-5', 'text-caption/4')).toBe('text-caption/4');
        expect(mergeDoodleClasses('text-caption/4', 'leading-5')).toBe('text-caption/4 leading-5');
    });

    it('continues merging colors independently of typography', () => {
        expect(mergeDoodleClasses('text-body-2 text-text-main', 'text-text-muted')).toBe('text-body-2 text-text-muted');
    });
});
