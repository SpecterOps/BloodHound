import { Theme } from '@mui/material';
import { MultiDirectedGraph } from 'graphology';
import * as layoutDagre from 'src/hooks/useLayoutDagre/useLayoutDagre';
import { initGraph } from './utils';

const layoutDagreSpy = vi.spyOn(layoutDagre, 'layoutDagre');

describe('Explore utils', () => {
    describe('initGraph', () => {
        const mockTheme = {
            palette: {
                color: { primary: '', links: '' },
                neutral: { primary: '', secondary: '' },
                common: { black: '', white: '' },
            },
        };
        it('calls sequentialLayout as the default graph layout', () => {
            const graph = new MultiDirectedGraph();
            initGraph(graph, { nodes: {}, edges: [] }, mockTheme as Theme, false);

            expect(layoutDagreSpy).toBeCalled();
        });
    });
});
