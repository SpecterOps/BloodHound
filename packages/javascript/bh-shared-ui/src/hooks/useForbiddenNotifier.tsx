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

import { useNotifications } from '../providers';
import { Permission } from '../utils';
import { useOnMount } from './useOnMount';

export const useForbiddenNotifier = (need: Permission, have: Permission[], message: string, key: string): boolean => {
    const { addNotification, dismissNotification } = useNotifications();
    const hasPermission = !!have?.includes(need);
    const effect = () => {
        if (!hasPermission) {
            addNotification(`${message} Please contact your admnistrator for details.`, key, {
                persist: true,
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });
        }
    };
    const cleanup = () => {
        dismissNotification(key);
    };

    useOnMount(effect, cleanup);

    return !hasPermission;
};
