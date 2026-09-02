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

import { faker } from '@faker-js/faker/locale/en';
import { GraphKindsResponse } from 'js-client-library';

const createEdgeKinds = (): GraphKindsResponse['data']['kinds'] => {
    return faker.helpers.uniqueArray(faker.random.word, 10);
};

const createNodeKinds = (): GraphKindsResponse['data']['kinds'] => {
    return faker.helpers.uniqueArray(faker.random.word, 10);
};

export const createGraphKinds = (nodeKinds?: string[], edgeKinds?: string[]): GraphKindsResponse['data'] => {
    const node_kinds = nodeKinds ?? createNodeKinds();
    const edge_kinds = edgeKinds ?? createEdgeKinds();

    return { kinds: [...edge_kinds, ...node_kinds] };
};
