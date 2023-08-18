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

import { createContext, Dispatch, ReactNode, useReducer } from 'react';
import { SnackbarKey, SnackbarProvider, useSnackbar } from 'notistack';
import { Notification } from './model';
import { NotificationAction } from './actions';
import { notificationsReducer } from './reducer';
import { IconButton, SvgIcon } from '@mui/material';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTimes } from '@fortawesome/free-solid-svg-icons';

export const NotificationsContext = createContext<Notification[]>([]);
export const NotificationsDispatchContext = createContext<Dispatch<NotificationAction> | null>(null);

const useDismissAction = (key: SnackbarKey) => {
    const { closeSnackbar } = useSnackbar();
    return (
        <IconButton size='small' color='inherit' onClick={() => closeSnackbar(key)}>
            <SvgIcon>
                <FontAwesomeIcon icon={faTimes} />
            </SvgIcon>
        </IconButton>
    );
};

interface NotificationProviderProps {
    children?: ReactNode;
}

const NotificationsProvider = ({ children }: NotificationProviderProps) => {
    const [notifications, dispatch] = useReducer(notificationsReducer, []);

    return (
        <NotificationsContext.Provider value={notifications}>
            <NotificationsDispatchContext.Provider value={dispatch}>
                <SnackbarProvider action={useDismissAction}>{children}</SnackbarProvider>
            </NotificationsDispatchContext.Provider>
        </NotificationsContext.Provider>
    );
};

export default NotificationsProvider;
