import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import * as exportUtils from './exportGraphData';

describe('exportGraphData', () => {
    beforeEach(() => {
        vi.useFakeTimers();
        // Filename timestamps use UTC (via DateTime.utc() + LuxonFormat.DATETIME_FILESYSTEM_SAFE),
        // so pin a UTC instant.
        // 2026-02-14T20:27:20.000Z → expected suffix: 2026-02-14_20-27-20
        vi.setSystemTime(new Date('2026-02-14T20:27:20.000Z'));
    });

    afterEach(() => {
        vi.useRealTimers();
        vi.restoreAllMocks();
        vi.unstubAllGlobals();
    });

    it('generates a human-friendly default filename for graph exports', () => {
        const name = exportUtils.getDefaultGraphExportFileName();

        expect(name).toBe('bh-graph-2026-02-14_20-27-20.json');
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
        expect(createdAnchor?.download).toBe('bh-graph-2026-02-14_20-27-20.json');
        expect(createObjectUrlSpy).toHaveBeenCalled();
    });
});