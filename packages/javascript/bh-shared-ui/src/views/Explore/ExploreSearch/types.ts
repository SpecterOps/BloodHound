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

import { SearchResult } from '../../../hooks/useSearch';
import { EntityKinds } from '../../../utils/content';

export interface SearchNodeType {
    objectid: string;
    type?: EntityKinds;
    name?: string;
}

//The search value usually aligns with the results from hitting the search endpoint but when
//we are pulling the data from a different page and filling out the value ourselves it might
//not conform to our expected type
export type SearchValue = SearchNodeType | SearchResult;
