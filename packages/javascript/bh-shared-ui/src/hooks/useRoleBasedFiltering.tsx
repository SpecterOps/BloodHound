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

import { isETACRole } from '../utils/roles';
import { useSelf } from './useBloodHoundUsers';
import { useListDisplayRoles } from './useListDisplayRoles/useListDisplayRoles';

// Matches against the current logged in user's role and checks whether user has access to all environments and is an ETAC enabled role
const useRoleBasedFiltering = (): boolean => {
    const { data: self } = useSelf();
    const { data: roles } = useListDisplayRoles();
    const userRoleIds = self?.roles?.map((item) => item.id as number) ?? [];
    const selectedETACEnabledRole = userRoleIds.some((roleId) => isETACRole(roleId, roles));

    return self?.all_environments === false && selectedETACEnabledRole;
};

export default useRoleBasedFiltering;
