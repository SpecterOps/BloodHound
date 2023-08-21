// Copyright 2023 Specter Ops, Inc.
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

import { SelectedEdge } from 'bh-shared-ui';
import { render, screen } from 'src/test-utils';
import EdgeInfoContent from 'src/views/Explore/EdgeInfo/EdgeInfoContent';

describe('EdgeInfoContent', () => {
    test('Trying to view the edge info does not crash the app when selecting an unrecognized edge', () => {
        const selectedEdge: SelectedEdge = {
            id: '1',
            name: 'CustomEdge',
            data: { isACL: false },
            sourceNode: { data: { name: 'source node' } },
            targetNode: { data: { name: 'target node' } },
        };

        render(<EdgeInfoContent selectedEdge={selectedEdge} />);

        //The text is broken up into different elements because of the span that bolds the custom edge so these assertions are broken up to match that
        //The assertions use regex to avoid having to match on the white space
        expect(screen.getByText(/The edge/)).toBeInTheDocument();
        expect(screen.getByText(selectedEdge.name)).toBeInTheDocument();
        expect(
            screen.getByText(/does not have any additional contextual information at this time./)
        ).toBeInTheDocument();

        //The general object information is still available even though there is no contextual information available for the edge
        expect(screen.getByText(/Source Node:/)).toBeInTheDocument();
        expect(screen.getByText(/Target Node:/)).toBeInTheDocument();
        expect(screen.getByText(/Is ACL:/)).toBeInTheDocument();

        //The whole app does not crash and require a refresh
        expect(
            screen.queryByText('An unexpected error has occurred. Please refresh the page and try again.')
        ).not.toBeInTheDocument();
    });
});
