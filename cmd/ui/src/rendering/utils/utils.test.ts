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

import { getNodeRadius } from 'src/rendering/utils/utils';

describe('Getting the node radius for use in our rendering programs', () => {
    const inverseSqrtZoomRatio = 1;
    test('If for some reason the node size is not defined return 1 as a default', () => {
        const nodeSize = undefined;

        // The default is not dependent on whether the node is highlighted
        expect(getNodeRadius(true, inverseSqrtZoomRatio, nodeSize)).toBe(1);
        expect(getNodeRadius(false, inverseSqrtZoomRatio, nodeSize)).toBe(1);
    });

    test('The node radius is adjusted for nodes that have their "highlighted" property set to true since the highlight adds some padding', () => {
        const nodeSize = 10;

        expect(getNodeRadius(true, inverseSqrtZoomRatio, nodeSize)).toBeGreaterThan(
            getNodeRadius(false, inverseSqrtZoomRatio, nodeSize)
        );
    });

    test('The node radius is adjusted to take into account the current camera zoom ratio', () => {
        const nodeSize = 10;

        expect(getNodeRadius(true, 1, nodeSize)).not.toEqual(getNodeRadius(true, 0.5, nodeSize));
        expect(getNodeRadius(false, 1, nodeSize)).not.toEqual(getNodeRadius(false, 0.5, nodeSize));
    });
});
