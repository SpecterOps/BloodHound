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
import { useSnackbar } from 'notistack';
import React, { useCallback, useEffect } from 'react';
import { removeSnackbar } from 'src/ducks/global/actions';
import { useAppDispatch, useAppSelector } from 'src/store';

let displayed: string[] = [];

const Notifier: React.FC = () => {
    const dispatch = useAppDispatch();

    const { enqueueSnackbar, closeSnackbar } = useSnackbar();

    const notifications = useAppSelector((state) => state.global.view.notifications);

    const storeDisplayed = (id: string) => {
        displayed = [...displayed, id];
    };

    const removeDisplayed = (id: string) => {
        displayed = [...displayed.filter((key) => id !== key)];
    };

    const clickSnackbarDismiss = useCallback(
        (key: string) => {
            closeSnackbar(key);
        },
        [closeSnackbar]
    );

    const action = useCallback(
        (key: string) => (
            <IconButton size='small' color='inherit' onClick={() => clickSnackbarDismiss(key)}>
                <SvgIcon>
                    <FontAwesomeIcon icon={faTimes} />
                </SvgIcon>
            </IconButton>
        ),
        [clickSnackbarDismiss]
    );

    useEffect(() => {
        notifications.forEach(({ key, message, options = {}, dismissed = false }) => {
            if (dismissed) {
                closeSnackbar(key);
                return;
            }

            if (displayed.includes(key)) return;

            options = {
                ...options,
                action: action(key),
            };

            enqueueSnackbar(message, {
                key,
                ...options,
                onClose: (event, reason, myKey) => {
                    if (options.onClose) {
                        options.onClose(event, myKey, reason);
                    }
                },
                onExited: (event, myKey: string) => {
                    dispatch(removeSnackbar(myKey));
                    removeDisplayed(myKey);
                },
            });

            // keep track of snackbars that we've displayed
            storeDisplayed(key);
        });
    }, [notifications, closeSnackbar, enqueueSnackbar, dispatch, action]);

    return null;
};

Notifier.propTypes = {};
export default Notifier;
