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
import type { Meta, StoryObj } from '@storybook/react';
import MarkdownContent from './MarkdownContent';

const meta: Meta<typeof MarkdownContent> = {
    title: 'Components/MarkdownContent',
    component: MarkdownContent,
    tags: ['autodocs'],
    decorators: [
        (Story) => (
            <div className='max-w-3xl'>
                <Story />
            </div>
        ),
    ],
    args: {
        markdown: `# Collection guide

Use **SharpHound** to collect directory data. Review the *collection settings* before starting.

## Before you begin

- Select an environment.
- Confirm the collection account.
- Upload the resulting JSON files.

1. Collect data.
2. Review results.

> Only collect from environments you manage.

See the [BloodHound documentation](https://bloodhound.specterops.io/) for more details.

---

Use \`objectid\` to identify an object.

\`\`\`json
{ "name": "EXAMPLE.COM", "kind": "Domain" }
\`\`\`
`,
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const FormattedContent: Story = {};
export const GitHubFlavoredMarkdown: Story = {
    args: {
        markdown: `## Collection checklist

- [x] Configure environment
- [ ] Upload data

| Environment | Status |
| --- | --- |
| Production | Complete |
| Test | Pending |

~~Deprecated setting~~
`,
    },
};
export const Empty: Story = { args: { markdown: '' } };
