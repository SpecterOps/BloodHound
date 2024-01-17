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

export const HIGHLIGHTED_LABEL_BACKGROUND_COLOR = '#FF6D66';
export const HIGHLIGHTED_LABEL_FONT_COLOR = '#FFF';

export const STARTING_ZOOM_FADE_RATIO = 0.5;
export const ENDING_ZOOM_FADE_RATIO = 0.3;

export const calculateLabelOpacity = (inverseSqrtZoomRatio: number): number => {
    if (inverseSqrtZoomRatio >= STARTING_ZOOM_FADE_RATIO) {
        return 1;
    }

    if (inverseSqrtZoomRatio < STARTING_ZOOM_FADE_RATIO && inverseSqrtZoomRatio > ENDING_ZOOM_FADE_RATIO) {
        return (inverseSqrtZoomRatio - ENDING_ZOOM_FADE_RATIO) / (STARTING_ZOOM_FADE_RATIO - ENDING_ZOOM_FADE_RATIO);
    }

    return 0;
};

export const getNodeRadius = (highlighted: boolean, inverseSqrtZoomRatio: number, size?: number): number => {
    const PADDING = 2;
    let radius = 1;

    if (!size) return radius;

    if (highlighted) radius = size * inverseSqrtZoomRatio + (PADDING + 2) * inverseSqrtZoomRatio;
    else radius = size * inverseSqrtZoomRatio;

    return radius;
};
