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

import { Navigate, useLocation } from 'react-router-dom';
import { authExpiredSelector } from 'src/ducks/auth/authSlice';
import { ROUTE_EXPIRED_PASSWORD, ROUTE_LOGIN } from 'src/ducks/global/routes';
import { useAppSelector } from 'src/store';

const AuthenticatedRoute: React.FC<{ children: any }> = ({ children }): React.ReactElement => {
    const authState = useAppSelector((state) => state.auth);
    const isAuthExpired = useAppSelector(authExpiredSelector);
    const location = useLocation();

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

    return children;
};

export default AuthenticatedRoute;
