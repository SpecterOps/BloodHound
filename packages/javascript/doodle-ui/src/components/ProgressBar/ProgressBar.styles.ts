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
export const progressBarRootClasses = 'w-full';

export const progressBarTrackClasses = 'relative h-2 w-full overflow-hidden bg-neutral-200 dark:bg-gray-700';

export const progressBarIndicatorClasses =
    'h-full bg-primary motion-safe:transition-[width] data-[indeterminate]:w-1/3 motion-safe:data-[indeterminate]:animate-progress-indeterminate';
