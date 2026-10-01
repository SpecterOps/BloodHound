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
import { fn } from '@storybook/test';
import { AppIcon } from './AppIcon';
import ConfirmationDialog from './ConfirmationDialog';

const meta: Meta<typeof ConfirmationDialog> = {
    title: 'Components/ConfirmationDialog',
    component: ConfirmationDialog,
    tags: ['autodocs'],
    parameters: { layout: 'centered' },
    argTypes: {
        cancelIcon: { control: false },
        confirmIcon: { control: false },
        text: { control: 'text' },
    },
    args: {
        open: true,
        title: 'Confirm action',
        text: 'Do you want to continue?',
        onCancel: fn(),
        onConfirm: fn(),
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const ChallengeText: Story = {
    args: {
        title: 'Delete item',
        text: 'This action cannot be undone.',
        challengeText: 'DELETE',
        confirmText: 'Delete',
    },
};

export const Loading: Story = {
    args: {
        isLoading: true,
    },
};

export const Error: Story = {
    args: {
        error: 'The action could not be completed. Please try again.',
    },
};

export const CustomActions: Story = {
    args: {
        cancelText: 'Go back',
        confirmText: 'Continue',
        cancelIcon: <AppIcon.CaretDown size={16} />,
        confirmIcon: <AppIcon.Checkmark size={16} />,
        iconPosition: 'right',
    },
};
