import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import * as exportUtils from './exportGraphData';
import { DateTime } from 'luxon';
import { LuxonFormat } from './datetime';

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

        expect(name).toBe(`bh-graph-${DateTime.fromJSDate(fakeTime).toLocal().toFormat(LuxonFormat.DATETIME_FILESYSTEM_SAFE)}.json`);
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
        expect(createdAnchor?.download).toBe(`bh-graph-${DateTime.fromJSDate(fakeTime).toLocal().toFormat(LuxonFormat.DATETIME_FILESYSTEM_SAFE)}.json`);
        expect(createObjectUrlSpy).toHaveBeenCalled();
    });
});