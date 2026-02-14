import { describe, expect, it, vi } from 'vitest';
import * as exportUtils from './exportGraphData';

describe('exportGraphData', () => {
    it('generates a human-friendly default filename for graph exports', () => {
        vi.useFakeTimers();
        // 2026-02-14T20:27:20.000Z
        vi.setSystemTime(new Date('2026-02-14T20:27:20.000Z'));

        const name = exportUtils.getDefaultGraphExportFileName();

        expect(name).toBe('bh-graph-2026-02-14_20-27-20.json');

        vi.useRealTimers();
    });

    it('uses the default filename generator when exporting JSON', () => {
        vi.useFakeTimers();
        vi.setSystemTime(new Date('2026-02-14T20:27:20.000Z'));

        const originalCreateObjectURL = (window.URL as any).createObjectURL;
        const createObjectUrlSpy = vi.fn(() => 'blob:mock');
        Object.defineProperty(window.URL, 'createObjectURL', {
            value: createObjectUrlSpy,
            writable: true,
        });

        let createdAnchor: HTMLAnchorElement | undefined;
        const originalCreateElement = document.createElement.bind(document);
        const createElementSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName: any, options?: any) => {
            const el = originalCreateElement(tagName, options) as any;
            if (tagName === 'a') createdAnchor = el as HTMLAnchorElement;
            return el;
        });

        const originalMouseEvent = globalThis.MouseEvent;
        class MockMouseEvent extends Event {
            constructor(type: string, _init?: any) {
                super(type);
            }
        }
        Object.defineProperty(globalThis, 'MouseEvent', {
            value: MockMouseEvent,
            writable: true,
        });

        exportUtils.exportToJson({ a: 1 });

        expect(createdAnchor?.download).toBe('bh-graph-2026-02-14_20-27-20.json');
        expect(createObjectUrlSpy).toHaveBeenCalled();

        createElementSpy.mockRestore();
        Object.defineProperty(window.URL, 'createObjectURL', {
            value: originalCreateObjectURL,
            writable: true,
        });
        Object.defineProperty(globalThis, 'MouseEvent', {
            value: originalMouseEvent,
            writable: true,
        });
        vi.useRealTimers();
    });
});
