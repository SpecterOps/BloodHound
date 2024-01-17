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

import { Settings } from 'sigma/settings';
import { NodeDisplayData, PartialButFor } from 'sigma/types';
import { calculateLabelOpacity } from '../utils/utils';

export default function drawLabel(
    context: CanvasRenderingContext2D,
    data: PartialButFor<NodeDisplayData & { inverseSqrtZoomRatio: number }, 'x' | 'y' | 'size' | 'label' | 'color'>,
    settings: Settings
): void {
    if (!data.label) return;
    const inverseSqrtZoomRatio = data.inverseSqrtZoomRatio || 1;

    const size = settings.labelSize * inverseSqrtZoomRatio,
        font = settings.labelFont,
        weight = settings.labelWeight,
        color = settings.labelColor.attribute
            ? data[settings.labelColor.attribute] || settings.labelColor.color || '#000'
            : settings.labelColor.color;

    context.globalAlpha = calculateLabelOpacity(inverseSqrtZoomRatio);

    context.fillStyle = color;
    context.font = `${weight} ${size}px ${font}`;

    context.fillText(
        data.label,
        data.x + data.size * inverseSqrtZoomRatio + 3 * inverseSqrtZoomRatio,
        data.y + size / 3
    );

    context.globalAlpha = 1;
}
