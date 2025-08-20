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

import type { OptionsObject } from 'notistack';
import { useContext, useMemo } from 'react';
import { NotificationsContext, NotificationsDispatchContext } from './NotificationsProvider';
import {
    addNotification as addNotificationAction,
    dismissNotification as dismissNotificationAction,
    removeNotification as removeNotificationAction,
    type NotificationAction,
} from './actions';

export const PERSIST_NOTIFICATION: OptionsObject = {
    persist: true,
    anchorOrigin: { vertical: 'top', horizontal: 'right' },
};

/** Make method that wraps an action creator with dispatch */
const curryWithDispatch = (dispatch: React.Dispatch<NotificationAction> | null) => {
    return <T extends (...args: any[]) => any>(actionCreator: T) => {
        return (...args: Parameters<T>) => {
            if (dispatch) {
                dispatch(actionCreator(...args));
            }
        };
    };
};

export const useNotifications = () => {
    const notifications = useContext(NotificationsContext);
    const dispatch = useContext(NotificationsDispatchContext);

    const actions = useMemo(() => {
        const withDispatch = curryWithDispatch(dispatch);
        return {
            addNotification: withDispatch(addNotificationAction),
            dismissNotification: withDispatch(dismissNotificationAction),
            removeNotification: withDispatch(removeNotificationAction),
        };
    }, [dispatch]);

    return {
        notifications,
        ...actions,
    };
};
