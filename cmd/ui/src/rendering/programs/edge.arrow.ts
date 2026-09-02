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
 * Sigma.js WebGL Renderer Edge Arrow Program
 * ===========================================
 *
 * Compound program rendering edges as an arrow from the source to the target.
 * Uses custom arrowhead program that is larger than sigma default.
 * @module
 */
import { createEdgeCompoundProgram } from 'sigma/rendering/webgl/programs/common/edge';
import EdgeArrowHeadProgram from 'src/rendering/programs/edge.arrowHead';
import EdgeClampedProgram from 'src/rendering/programs/edge.clamped';

const EdgeArrowProgram = createEdgeCompoundProgram([EdgeClampedProgram, EdgeArrowHeadProgram]);

export default EdgeArrowProgram;
