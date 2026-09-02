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
import { DateTime } from 'luxon';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { LuxonFormat } from './datetime';
import * as exportUtils from './exportGraphData';

const fakeTime = new Date('2026-02-14T20:27:20.000Z');

describe('exportGraphData', () => {
    const originalMouseEvent = globalThis.MouseEvent;

    beforeEach(() => {
        vi.useFakeTimers();
        vi.setSystemTime(fakeTime);
        class MockMouseEvent extends Event {
            constructor(type: string) {
                super(type);
            }
        }
        vi.stubGlobal('MouseEvent', MockMouseEvent);
    });

    afterEach(() => {
        vi.useRealTimers();
        vi.restoreAllMocks();
        globalThis.MouseEvent = originalMouseEvent;
    });

    it('generates a human-friendly default filename for graph exports', () => {
        const name = exportUtils.getDefaultGraphExportFileName();

        expect(name).toBe(
            `bh-graph-${DateTime.fromJSDate(fakeTime).toLocal().toFormat(LuxonFormat.DATETIME_FILESYSTEM_SAFE)}.json`
        );
    });

    it('uses the default filename generator when exporting JSON', () => {
        const createObjectUrlSpy = vi.spyOn(window.URL, 'createObjectURL').mockReturnValue('blob:mock');

        class MockMouseEvent extends Event {
            constructor(type: string) {
                super(type);
            }
        }
        vi.stubGlobal('MouseEvent', MockMouseEvent);

        let createdAnchor: HTMLAnchorElement | undefined;
        const originalCreateElement = document.createElement.bind(document);
        vi.spyOn(document, 'createElement').mockImplementation((tagName: any, options?: any) => {
            const el = originalCreateElement(tagName, options) as any;
            if (tagName === 'a') createdAnchor = el as HTMLAnchorElement;
            return el;
        });

        exportUtils.exportToJson({ a: 1 });

        expect(createdAnchor).toBeInstanceOf(HTMLAnchorElement);
        expect(createdAnchor?.download).toBe(
            `bh-graph-${DateTime.fromJSDate(fakeTime).toLocal().toFormat(LuxonFormat.DATETIME_FILESYSTEM_SAFE)}.json`
        );
        expect(createObjectUrlSpy).toHaveBeenCalled();
    });
});
