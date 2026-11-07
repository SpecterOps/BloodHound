// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import { cn } from '../utils/theme';

/** Temporary compatibility styling for legacy MUI tables. */
export const LEGACY_MUI_TABLE_CLASS = cn(
    'border border-solid border-neutral-light-4 !bg-neutral-1 !shadow-none dark:border-neutral-900',
    '[&_.MuiTableBody-root]:!bg-neutral-1',
    '[&_.MuiTableRow-root:not(.Mui-selected)]:!bg-neutral-1',
    '[&_.MuiTableRow-hover:not(.Mui-selected):hover]:!bg-neutral-3'
);
