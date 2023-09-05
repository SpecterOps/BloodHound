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

import { useSnackbar } from 'notistack';
import { useEffect } from 'react';
import { useNotifications } from '../providers';

let displayedNotifications: string[] = [];

const AppNotifications = () => {
    const { notifications, removeNotification } = useNotifications();
    const { enqueueSnackbar, closeSnackbar } = useSnackbar();

    const storeDisplayed = (id: string) => {
        displayedNotifications = [...displayedNotifications, id];
    };

    const removeDisplayed = (id: string) => {
        displayedNotifications = [...displayedNotifications.filter((key) => id !== key)];
    };

    useEffect(() => {
        notifications.forEach(({ key, message, options = {}, dismissed = false }) => {
            if (dismissed) {
                closeSnackbar(key);
            } else if (!displayedNotifications.includes(key)) {
                enqueueSnackbar(message, {
                    key,
                    ...options,
                    onClose: (event, reason, id) => {
                        if (options.onClose) {
                            options.onClose(event, reason, id);
                        }
                    },
                    onExited: (_, id) => {
                        removeNotification(id as string);
                        removeDisplayed(id as string);
                    },
                });
                storeDisplayed(key);
            }
        });
    }, [notifications, closeSnackbar, enqueueSnackbar, removeNotification]);

    return null;
};

export default AppNotifications;
