import { NodeProgramType } from 'sigma/rendering';

import createNodeGlyphProgram from './factory';

export { default as createNodeImageProgram } from './factory';
export const NodeImageProgram: NodeProgramType = createNodeGlyphProgram();
