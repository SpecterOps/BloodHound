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
import { faEyeSlash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Badge } from 'doodle-ui';
import useRoleBasedFiltering from '../../hooks/useRoleBasedFiltering';
import { useIsEnterprise } from '../../providers/AppNameProvider';

const useShowAccessDeniedBanner = () => {
    const userRoleNotAllowed = useRoleBasedFiltering();
    const isEnterprise = useIsEnterprise();

    return userRoleNotAllowed && isEnterprise;
};

export default function RoleBasedFilterBadge() {
    const showAccessDeniedBanner = useShowAccessDeniedBanner();

    if (showAccessDeniedBanner)
        return (
            <Badge
                data-testid='explore_entity-information-panel-role-based-filtering-badge'
                variant='fill'
                className='px-2 py-1'
                color='primary'
                icon={<FontAwesomeIcon icon={faEyeSlash} className='ml-1 mr-3' />}
                label='Role-based access filtering applied'
            />
        );
}
