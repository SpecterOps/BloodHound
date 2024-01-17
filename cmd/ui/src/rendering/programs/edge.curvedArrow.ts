// Copyright 2023 Specter Ops, Inc.
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

/**
 * Sigma.js WebGL Renderer Edge Curved Arrow Program
 * ===========================================
 *
 * Compound program rendering edges as a curved arrow from the source to the target.
 * Spreads curves out based on their position in the group.
 * @module
 */
import { createEdgeCompoundProgram } from 'sigma/rendering/webgl/programs/common/edge';
import EdgeCurvedArrowHeadProgram from './edge.curvedArrowHead';
import EdgeCurvedProgram from './edge.curved';
import { EdgeDisplayData, Coordinates } from 'sigma/types';
import { EdgeDirection } from 'src/utils';

export type CurvedEdgeDisplayData = EdgeDisplayData & {
    groupSize?: number;
    groupPosition?: number;
    direction?: EdgeDirection;
    control?: Coordinates;
    inverseSqrtZoomRatio?: number;
};

const EdgeArrowProgram = createEdgeCompoundProgram([EdgeCurvedProgram, EdgeCurvedArrowHeadProgram]);

export default EdgeArrowProgram;
