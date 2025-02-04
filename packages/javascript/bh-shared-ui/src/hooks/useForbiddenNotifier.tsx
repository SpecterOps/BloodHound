// Copyright 2025 Specter Ops, Inc.
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

import { useEffect } from 'react';
import { useNotifications } from '../providers';
import { Permission } from '../utils';

export const useForbiddenNotifier = (need: Permission, have: Permission[], message: string, key: string): boolean => {
    const { addNotification, dismissNotification } = useNotifications();
    const hasPermission = !!have?.includes(need);

    useEffect(() => {
        if (!hasPermission) {
            addNotification(`${message} Please contact your admnistrator for details.`, key, {
                persist: true,
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });
        }

        return () => {
            dismissNotification(key);
        };
        // This linting is disabled because adding the dependencies would cause a render loop
        // and we only want the effect to happen on first render.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return !hasPermission;
};
