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

import { CreateSelectorRequest, SelectorSeedRequest, UpdateSelectorRequest } from 'js-client-library/dist/requests';
import { ISO_DATE_STRING } from '../../../../utils';

export interface SelectorFormInputs {
    name: string;
    description: string;
    autoCertify: boolean;
    disabled_at: ISO_DATE_STRING | null;
    seeds: SelectorSeedRequest[];
}

export interface CreateSelectorParams {
    tagId: string | number;
    values: CreateSelectorRequest;
}
export interface DeleteSelectorParams {
    tagId: string | number;
    selectorId: string | number;
}

export interface PatchSelectorParams extends DeleteSelectorParams {
    updatedValues: UpdateSelectorRequest;
}
