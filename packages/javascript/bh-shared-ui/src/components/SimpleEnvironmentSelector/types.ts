// Copyright 2024 Specter Ops, Inc.
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

import { Environment } from 'js-client-library';

export const AD_PLATFORM = 'active-directory-platform' as const;
export const AZ_PLATFORM = 'azure-platform' as const;

export type EnvironmentPlatform = typeof AD_PLATFORM | typeof AZ_PLATFORM;
export type SelectorValueTypes = Environment['type'] | EnvironmentPlatform;

export type SelectedEnvironment = { type: SelectorValueTypes | null; id: string | null };
