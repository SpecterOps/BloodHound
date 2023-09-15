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

import { Notification } from './model';

export enum ActionType {
    Add = 'add',
    Dismiss = 'dismiss',
    Remove = 'remove',
}

export type Add = { type: ActionType.Add; value: Notification };
export type Dismiss = { type: ActionType.Dismiss; key?: string };
export type Remove = { type: ActionType.Remove; key?: string };
export type NotificationAction = Add | Dismiss | Remove;

export const addNotification = (notification: string, key?: string, options: any = {}): Add => {
    return {
        type: ActionType.Add,
        value: {
            message: notification,
            key: key || (new Date().getTime() + Math.random()).toString(),
            options: {
                ...options,
                autoHideDuration: 5000,
            },
            dismissed: false,
        },
    };
};

export const dismissNotification = (key?: string): Dismiss => {
    return {
        type: ActionType.Dismiss,
        key: key,
    };
};

export const removeNotification = (key?: string): Remove => {
    return {
        type: ActionType.Remove,
        key: key,
    };
};
