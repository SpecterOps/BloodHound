import { describe, expect, it, vi } from 'vitest';
import * as exportUtils from './exportGraphData';

describe('exportGraphData', () => {
    it('generates a human-friendly default filename for graph exports', () => {
        vi.useFakeTimers();
        // 2026-02-14T20:27:20.000Z
        vi.setSystemTime(new Date('2026-02-14T20:27:20.000Z'));

        // This function will be added in the next task
        // @ts-expect-error - intentional until implemented
        const name = exportUtils.getDefaultGraphExportFileName();

        expect(name).toBe('bh-graph-2026-02-14_20-27-20.json');

        vi.useRealTimers();
    });
});
