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

import { faTimes } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton, SvgIcon } from '@mui/material';
import { Alert } from 'doodle-ui';
import { SnackbarContent, SnackbarKey, SnackbarProvider, useSnackbar } from 'notistack';
import React, { Dispatch, ReactNode, createContext, useReducer } from 'react';
import { NotificationAction } from './actions';
import { Notification } from './model';
import { notificationsReducer } from './reducer';
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

type SnackVariant = 'default' | 'error' | 'info' | 'warning' | 'success';

interface MyCustomSnackProps {
    id: string | number;
    message: string | React.ReactNode;
    variant: SnackVariant | null | undefined;
    title?: string;
}

export const MyCustomSnack = React.forwardRef<HTMLDivElement, MyCustomSnackProps>(
    ({ id, message, variant, title }, ref) => {
        const { closeSnackbar } = useSnackbar();
        return (
            <SnackbarContent ref={ref} style={{ justifyContent: 'center' }}>
                <Alert variant={variant} title={title} onClose={() => closeSnackbar(id)}>
                    {message}
                </Alert>
            </SnackbarContent>
        );
    }
);

MyCustomSnack.displayName = 'MyCustomSnack';

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
