// Copyright 2025 Specter Ops, Inc.
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

import * as IconOptions from './Icons';

export type AppIconOptions = keyof typeof IconOptions;
/**
 * Want to add an icon? Follow these steps:
 *
 * 1. Create a new file under the `Icons/` directory with the name of your icon. *I.e. Simulation*
 * 2. Create a __named export__ react component that receives `BaseSVGProps` and returns the icons svg
 * 3. Swap out the svg and path elements with the `BaseSVG` and `BasePath` components. Pass all props from parent to `BaseSVG`
 * 4. Implementation `<AppIcon.Simulation />`
 */
export const AppIcon = IconOptions;
