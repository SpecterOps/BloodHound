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

import { Routable } from 'bh-shared-ui';
import { Badge } from 'doodle-ui';
import { useMemo } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { authExpiredSelector } from 'src/ducks/auth/authSlice';
import { ROUTE_EXPIRED_PASSWORD, ROUTE_LOGIN } from 'src/routes/constants';
import { useAppSelector } from 'src/store';

type AuthenticatedRouteProps = {
    children: React.ReactElement;
    disallowedRoles?: Routable['disallowedRoles'];
};

const AuthenticatedRoute: React.FC<AuthenticatedRouteProps> = ({ children, disallowedRoles = [] }) => {
    const authState = useAppSelector((state) => state.auth);
    const isAuthExpired = useAppSelector(authExpiredSelector);
    const location = useLocation();

    const invalidRolesForThisUser = useMemo(
        () =>
            disallowedRoles?.filter((disallowedRole) =>
                authState.user?.roles.find((role: any) => role.name === disallowedRole.name)
            ),
        [authState?.user, disallowedRoles]
    );

    // If user is not authenticated, redirect to login screen
    if (authState.sessionToken === null || authState.user === null) {
        return <Navigate to={ROUTE_LOGIN} state={{ from: location }} />;
    }

    // If user password is expired, redirect to expired password screen unless they are on the expired password screen
    if (isAuthExpired) {
        if (location.pathname !== ROUTE_EXPIRED_PASSWORD) {
            return <Navigate to={ROUTE_EXPIRED_PASSWORD} state={{ from: location }} />;
        } else {
            return children;
        }
    }

    const hasDisallowedRoles = invalidRolesForThisUser.length > 0;

    if (hasDisallowedRoles) {
        const label = invalidRolesForThisUser?.[0].notification;

        return (
            <Badge
                data-testid='attack-paths_etac-filtering-badge'
                variant='fill'
                className='px-2 py-1 ml-auto absolute right-4 top-4'
                color='red'
                label={label}
            />
        );
    }

    return children;
};

export default AuthenticatedRoute;
