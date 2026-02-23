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

import { Tooltip } from '@bloodhoundenterprise/doodleui';
import { ReactNode } from 'react';

type Props = {
    children: ReactNode;
    condition: boolean;
    tooltip: string;
};

/**
 * Conditionally wraps children with a Tooltip component based on the condition prop.
 * If condition is true, renders children wrapped in a Tooltip. Otherwise, renders children directly.
 *
 * @param condition - Whether to show the tooltip
 * @param tooltip - The tooltip text to display
 * @param children - The child elements to render
 */
export function ConditionalTooltip({ condition, tooltip, children }: Props) {
    return condition ? <Tooltip tooltip={tooltip}>{children}</Tooltip> : children;
}
