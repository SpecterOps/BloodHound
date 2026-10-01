// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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
import { faStar } from '@fortawesome/free-solid-svg-icons';
import type { Meta, StoryObj } from '@storybook/react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../graphSchema';
import { customNodeKindsKeys } from '../../hooks/useCustomNodeKinds';
import { NODE_ICONS } from '../../utils/icons';
import NodeIcon from './NodeIcon';

const customNodeKind = 'ExampleCustomNode';
const nodeTypes = Object.keys(NODE_ICONS);
const queryClient = new QueryClient();
queryClient.setQueryData(customNodeKindsKeys.all, {
    [customNodeKind]: { icon: faStar, color: '#BDA8F0' },
});

const iconMarkup = (nodeType: string) => `<NodeIcon nodeType=${JSON.stringify(nodeType)} />`;

const NodeIconGallery = () => (
    <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'>
        {nodeTypes.map((nodeType) => (
            <div
                key={nodeType}
                className='min-w-0 flex flex-col items-center gap-3 rounded-lg border border-neutral-3 bg-neutral-1 p-4'>
                <div className='flex h-20 w-full items-center justify-center'>
                    <NodeIcon nodeType={nodeType} />
                </div>
                <span className='w-full break-words text-center text-sm font-semibold'>{nodeType}</span>
                <code className='w-full whitespace-pre-wrap break-words rounded bg-neutral-2 p-2 text-xs'>
                    {iconMarkup(nodeType)}
                </code>
            </div>
        ))}
    </div>
);

const meta = {
    title: 'Components/NodeIcon',
    component: NodeIcon,
    tags: ['autodocs'],
    parameters: { layout: 'centered' },
    decorators: [
        (Story) => (
            <QueryClientProvider client={queryClient}>
                <Story />
            </QueryClientProvider>
        ),
    ],
    args: { nodeType: ActiveDirectoryNodeKind.User },
    argTypes: {
        nodeType: {
            options: [
                ActiveDirectoryNodeKind.User,
                ActiveDirectoryNodeKind.Computer,
                ActiveDirectoryNodeKind.Group,
                AzureNodeKind.User,
                AzureNodeKind.VM,
                customNodeKind,
                'UnknownNode',
            ],
            control: 'select',
        },
        className: { control: 'text' },
    },
} satisfies Meta<typeof NodeIcon>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Azure: Story = {
    args: { nodeType: AzureNodeKind.User },
};

export const Custom: Story = {
    args: { nodeType: customNodeKind },
};

export const Unknown: Story = {
    args: { nodeType: 'UnknownNode' },
};

export const Gallery: Story = {
    render: NodeIconGallery,
    parameters: {
        layout: 'padded',
        controls: { disable: true },
        docs: {
            description: {
                story: 'Browse built-in node icons. Each card shows the node kind and JSX needed to render it.',
            },
            source: {
                type: 'code',
                code: `import { NodeIcon } from 'bh-shared-ui';

<>
${nodeTypes.map((nodeType) => `    ${iconMarkup(nodeType)}`).join('\n')}
</>`,
            },
        },
    },
};
