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

import { ActionType, NotificationAction } from './actions';
import { Notification } from './model';

export const notificationsReducer = (state: Notification[], action: NotificationAction): Notification[] => {
    if (action.type === ActionType.Add) {
        return [...state, action.value];
    } else if (action.type === ActionType.Dismiss) {
        return state.map((notification) => {
            return action.key === null || action.key === notification.key
                ? { ...notification, dismissed: true }
                : { ...notification };
        });
    } else {
        return state.filter((notification) => notification.key !== action.key);
    }
};
