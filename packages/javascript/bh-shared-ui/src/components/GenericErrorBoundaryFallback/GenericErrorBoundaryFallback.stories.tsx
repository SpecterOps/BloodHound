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
import { ErrorBoundary } from 'react-error-boundary';
import GenericErrorBoundaryFallback from './GenericErrorBoundaryFallback';

const meta: Meta<typeof GenericErrorBoundaryFallback> = {
    title: 'Components/GenericErrorBoundaryFallback',
    component: GenericErrorBoundaryFallback,
    tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
function FailedContent(): never {
    throw new Error('Intentional Storybook error to demonstrate the boundary fallback.');
}
function BoundaryExample() {
    return (
        <ErrorBoundary FallbackComponent={GenericErrorBoundaryFallback}>
            <FailedContent />
        </ErrorBoundary>
    );
}
export const CaughtError: Story = {
    render: () => <BoundaryExample />,
    parameters: {
        docs: {
            description: {
                story: 'Intentionally throws a render error, which the real error boundary catches. The error is also reported in the browser console.',
            },
        },
    },
};
