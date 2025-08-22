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
export enum CustomRangeError {
    INVALID_DATE = 'Must be a valid date in yyyy-mm-dd format.',
    INVALID_RANGE_START = 'Start date must be before end date.',
    INVALID_RANGE_END = 'End date must be after start date.',
}

export const START_DATE = 'start-date' as const;
export const END_DATE = 'end-date' as const;
