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

import userEvent from '@testing-library/user-event';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../graphSchema';
import { render, screen } from '../../test-utils';
import { BasicObjectInfoFields } from './BasicObjectInfoFields';

const nodeName = 'test-node-name';

const relatedKindCases = [
    {
        description: 'Service Principal ID',
        property: 'service_principal_id',
        id: 'service-principal-id-value',
        label: 'Service Principal ID:',
        kind: AzureNodeKind.ServicePrincipal,
    },
    {
        description: 'Federated Identity Credential Application ID',
        property: 'federatedidentitycredentialappid',
        id: 'federated-identity-credential-app-id-value',
        label: 'Federated Identity Credential Application ID:',
        kind: AzureNodeKind.App,
    },
    {
        description: 'Node Resource Group ID',
        property: 'noderesourcegroupid',
        id: 'node-resource-group-id-value',
        label: 'Node Resource Group ID:',
        kind: AzureNodeKind.ResourceGroup,
    },
    {
        description: 'Linked Group ID',
        property: 'grouplinkid',
        id: 'group-link-id-value',
        label: 'Linked Group ID:',
        kind: ActiveDirectoryNodeKind.Group,
    },
] as const;

describe('BasicObjectInfoFields related kind fields', () => {
    describe.each(relatedKindCases)('$description', ({ property, id, label, kind }) => {
        it('renders the label and clickable id when handleSourceNodeSelected is provided', () => {
            render(
                <BasicObjectInfoFields
                    properties={{ name: nodeName, [property]: id }}
                    handleSourceNodeSelected={vi.fn()}
                />
            );

            expect(screen.getByText(label)).toBeInTheDocument();
            expect(screen.getByText(id)).toBeInTheDocument();
        });

        it('calls handleSourceNodeSelected with the related node when the id is clicked', async () => {
            const user = userEvent.setup();
            const handleSourceNodeSelected = vi.fn();

            render(
                <BasicObjectInfoFields
                    properties={{ name: nodeName, [property]: id }}
                    handleSourceNodeSelected={handleSourceNodeSelected}
                />
            );

            await user.click(screen.getByText(id));

            expect(handleSourceNodeSelected).toHaveBeenCalledTimes(1);
            expect(handleSourceNodeSelected).toHaveBeenCalledWith({
                objectid: id,
                type: kind,
                name: nodeName,
            });
        });

        it('passes an empty name to handleSourceNodeSelected when the node has no name', async () => {
            const user = userEvent.setup();
            const handleSourceNodeSelected = vi.fn();

            render(
                <BasicObjectInfoFields
                    properties={{ [property]: id }}
                    handleSourceNodeSelected={handleSourceNodeSelected}
                />
            );

            await user.click(screen.getByText(id));

            expect(handleSourceNodeSelected).toHaveBeenCalledWith({
                objectid: id,
                type: kind,
                name: '',
            });
        });

        it('does not render the field when handleSourceNodeSelected is not provided', () => {
            render(<BasicObjectInfoFields properties={{ name: nodeName, [property]: id }} />);

            expect(screen.queryByText(label)).not.toBeInTheDocument();
            expect(screen.queryByText(id)).not.toBeInTheDocument();
        });
    });

    it('renders every related kind field when all special properties are present', () => {
        const properties = relatedKindCases.reduce((acc, { property, id }) => ({ ...acc, [property]: id }), {
            name: nodeName,
        });

        render(<BasicObjectInfoFields properties={properties} handleSourceNodeSelected={vi.fn()} />);

        relatedKindCases.forEach(({ label, id }) => {
            expect(screen.getByText(label)).toBeInTheDocument();
            expect(screen.getByText(id)).toBeInTheDocument();
        });
    });
});
