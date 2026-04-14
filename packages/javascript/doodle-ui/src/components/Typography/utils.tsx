// Copyright 2026 Specter Ops, Inc.
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
export type Variant =
    | 'h1'
    | 'h2'
    | 'h3'
    | 'h4'
    | 'h5'
    | 'h6'
    | 'subtitle1'
    | 'subtitle2'
    | 'body1'
    | 'body2'
    | 'caption';

export const DEFAULT_VARIANT: Variant = 'body1';

export const variantMapping: Record<Variant, keyof JSX.IntrinsicElements> = {
    h1: 'h1',
    h2: 'h2',
    h3: 'h3',
    h4: 'h4',
    h5: 'h5',
    h6: 'h6',
    subtitle1: 'h6',
    subtitle2: 'h6',
    body1: 'p',
    body2: 'p',
    caption: 'span',
};

export const tagOptions = [
    undefined,
    // Headings — document outline hierarchy
    'h1',
    'h2',
    'h3',
    'h4',
    'h5',
    'h6',
    // Block elements — structural then textual
    'div',
    'p',
    'pre',
    // Inline semantic — meaning-bearing
    'code',
    'cite',
    'mark',
    'strong',
    'em',
    // Inline presentational — visual only
    'b',
    'i',
    'u',
    // Inline text modification
    'del',
    'ins',
    'sup',
    'sub',
    'small',
];
