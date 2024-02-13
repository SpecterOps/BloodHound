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

import React from 'react';

import { Navigate } from 'react-router-dom';
import { authExpiredSelector, logout, updateExpiredPassword } from 'src/ducks/auth/authSlice';
import LoginPage from 'src/components/LoginPage';
import PasswordResetForm from 'src/components/PasswordResetForm';
import { ROUTE_HOME } from 'src/ducks/global/routes';
import { useAppDispatch, useAppSelector } from 'src/store';

const PasswordReset: React.FC = () => {
    /* Hooks */
    const dispatch = useAppDispatch();
    const authState = useAppSelector((state) => state.auth);
    const authExpired = useAppSelector(authExpiredSelector);

    /* Event Handlers */
    const handleSubmit = (password: string) => {
        dispatch(updateExpiredPassword({ password }));
    };

    const handleCancel = () => {
        dispatch(logout());
    };

    // Redirect if auth is not expired
    if (authExpired === false) return <Navigate to={ROUTE_HOME} />;

    return (
        <LoginPage>
            <PasswordResetForm
                onSubmit={handleSubmit}
                onCancel={handleCancel}
                loading={authState.updateExpiredPasswordLoading}
            />
        </LoginPage>
    );
};

export default PasswordReset;
