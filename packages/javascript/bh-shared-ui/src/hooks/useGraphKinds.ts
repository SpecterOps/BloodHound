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
import { useQuery } from 'react-query';
import { graphSchema } from '../constants';
import { apiClient } from '../utils/api';

export const graphKindsKeys = {
    all: ['graph-kinds'] as const,
    nodes: () => [...graphKindsKeys.all, 'nodes'] as const,
    edges: () => [...graphKindsKeys.all, 'edges'] as const,
};

export const useGraphNodeKinds = () => {
    const params = new URLSearchParams({ type: 'eq:node' });
    return useQuery({
        queryKey: graphKindsKeys.nodes(),
        queryFn: ({ signal }) => apiClient.getKinds({ signal, params }).then((res) => res.data.data),
    });
};

const useGraphEdgeKinds = () => {
    const params = new URLSearchParams({ type: 'eq:edge' });
    return useQuery({
        queryKey: graphKindsKeys.edges(),
        queryFn: ({ signal }) => apiClient.getKinds({ signal, params }).then((res) => res.data.data),
    });
};

export const useCypherSchema = () => {
    const { data: nodeKinds } = useGraphNodeKinds();
    const { data: edgeKinds } = useGraphEdgeKinds();

    const schema = graphSchema({ nodes: nodeKinds?.kinds, edges: edgeKinds?.kinds });

    return schema;
};
