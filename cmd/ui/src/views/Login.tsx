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

import React, { useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import LoginForm from 'src/components/LoginForm';
import LoginViaSAMLForm from 'src/components/LoginViaSAMLForm';
import LoginPage from 'src/components/LoginPage';
import { useQuery, useQueryClient } from 'react-query';
import { apiClient } from 'bh-shared-ui';
import { Box, CircularProgress } from '@mui/material';
import { OneTimePasscodeForm } from 'bh-shared-ui';

import { login as loginAction, logout } from 'src/ducks/auth/authSlice';
import { ROUTE_HOME, ROUTE_USER_DISABLED } from 'src/ducks/global/routes';
import { useAppDispatch, useAppSelector } from 'src/store';

const Login: React.FC = () => {
    /* Hooks */
    const dispatch = useAppDispatch();

    const queryClient = useQueryClient();

    const authState = useAppSelector((state) => state.auth);

    const [useSAML, setUseSAML] = useState(false);

    const [lastUsername, setLastUsername] = useState('');

    const [lastPassword, setLastPassword] = useState('');

    // clear the react-query query cache between authenticated sessions
    useEffect(() => {
        queryClient.clear();
    }, [queryClient]);

    const listSAMLSignOnEndpointsQuery = useQuery(['listSAMLSignOnEndpoints'], ({ signal }) =>
        apiClient.listSAMLSignOnEndpoints({ signal }).then((res) => res.data.data.endpoints)
    );

    /* Event Handlers */
    const resetForm = () => {
        setUseSAML(false);
        setLastUsername('');
        setLastPassword('');
        dispatch(logout());
    };

    const handleSubmitLoginForm = async (username: string, password: string) => {
        setLastUsername(username);
        setLastPassword(password);
        dispatch(loginAction({ username, password }));
    };

    const handleSubmitLoginWithOneTimePasscodeForm = async (otp: string) => {
        dispatch(loginAction({ username: lastUsername, password: lastPassword, otp }));
    };

    const handleSubmitLoginViaSAMLForm = (redirectURL: string) => {
        window.location.assign(redirectURL);
    };

    /* Implementation */

    // Redirect if already logged in
    if (authState.sessionToken !== null && authState.user !== null) return <Navigate to={ROUTE_HOME} />;

    if (listSAMLSignOnEndpointsQuery.isLoading) {
        return (
            <LoginPage>
                <Box textAlign='center'>
                    <CircularProgress />
                </Box>
            </LoginPage>
        );
    }

    const userIsDisabled = authState.loginError?.response?.status === 403 || false;
    if (userIsDisabled) {
        return <Navigate to={ROUTE_USER_DISABLED} />;
    }

    const oneTimePasscodeRequired = authState.loginError?.response?.status === 400 || false;
    if (oneTimePasscodeRequired) {
        return (
            <LoginPage>
                <OneTimePasscodeForm
                    onSubmit={handleSubmitLoginWithOneTimePasscodeForm}
                    onCancel={resetForm}
                    loading={authState.loginLoading}
                />
            </LoginPage>
        );
    }

    if (listSAMLSignOnEndpointsQuery.isError || listSAMLSignOnEndpointsQuery.data?.length === 0) {
        return (
            <LoginPage>
                <LoginForm onSubmit={handleSubmitLoginForm} loading={authState.loginLoading} />
            </LoginPage>
        );
    }

    if (useSAML) {
        return (
            <LoginPage>
                <LoginViaSAMLForm
                    providers={listSAMLSignOnEndpointsQuery.data}
                    onSubmit={handleSubmitLoginViaSAMLForm}
                    onCancel={resetForm}
                />
            </LoginPage>
        );
    }

    return (
        <LoginPage>
            <LoginForm
                onSubmit={handleSubmitLoginForm}
                onLoginViaSAML={() => setUseSAML(true)}
                loading={authState.loginLoading}
            />
        </LoginPage>
    );
};

export default Login;
