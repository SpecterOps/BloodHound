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

import { createAsyncThunk, createSelector, createSlice } from '@reduxjs/toolkit';
import { DateTime } from 'luxon';
import { apiClient } from 'bh-shared-ui';
import type { AppDispatch, AppState } from 'src/store';
import { addSnackbar } from '../global/actions';
import * as types from './types';

export const initialState: types.AuthState = {
    isInitialized: false,
    loginLoading: false,
    loginError: null,
    updateExpiredPasswordLoading: false,
    updateExpiredPasswordError: null,
    sessionToken: null,
    user: null,
};

export const login = createAsyncThunk(
    'auth/login',
    async (arg: { username: string; password: string; otp?: string }, { dispatch, rejectWithValue, signal }) => {
        try {
            const loginResponse = await apiClient.login(
                {
                    login_method: 'secret',
                    secret: arg.password,
                    username: arg.username,
                    otp: arg.otp,
                },
                { signal }
            );
            const getSelfResponse = await apiClient.baseClient.get('/api/v2/self', {
                headers: {
                    Authorization: `Bearer ${loginResponse.data.data.session_token}`,
                },
                signal,
            });
            return {
                sessionToken: loginResponse.data.data.session_token,
                user: getSelfResponse.data.data,
            };
        } catch (error: any) {
            // one time passcode provided but error occurred
            if (error.response && error.response.status === 400 && arg.otp !== undefined) {
                dispatch(addSnackbar('Invalid token. Please try again.', 'loginError'));
                return rejectWithValue(error);
            }
            // one time passcode required but not provided
            else if (
                error.response &&
                (error.response.status === 400 || error.response.status === 403) &&
                arg.otp === undefined
            ) {
                return rejectWithValue(error);
            }
            // any other error
            else {
                dispatch(addSnackbar('Login failed. Please try again.', 'loginError'));
                return rejectWithValue(error);
            }
        }
    }
);

export const logout = createAsyncThunk('auth/logout', async () => {
    return await apiClient.logout().catch(() => {});
});

export const initialize = createAsyncThunk<
    types.getSelfResponse,
    void,
    {
        dispatch: AppDispatch;
        state: AppState;
    }
>('auth/initialize', async (_, { getState, rejectWithValue }) => {
    const sessionToken = getState().auth.sessionToken;
    if (sessionToken === null) {
        throw new Error('No session token provided');
    }
    try {
        const getSelfResponse = await apiClient.baseClient.get('/api/v2/self', {
            headers: {
                Authorization: `Bearer ${sessionToken}`,
            },
        });
        return getSelfResponse.data.data;
    } catch (error) {
        return rejectWithValue(error);
    }
});

export const updateExpiredPassword = createAsyncThunk<
    types.getSelfResponse,
    {
        password: string;
    },
    {
        dispatch: AppDispatch;
        state: AppState;
    }
>('auth/updateExpiredPassword', async ({ password }, { getState, dispatch, rejectWithValue }) => {
    const userId = getState().auth.user?.id;
    if (userId === undefined) {
        throw new Error('Could not find user ID');
    }

    try {
        await apiClient.putUserAuthSecret(userId, {
            needs_password_reset: false,
            secret: password,
        });
        const response = await apiClient.getSelf();
        return response.data.data;
    } catch (error) {
        dispatch(
            addSnackbar(
                'An error occurred when attempting to reset your password. Please try again.',
                'updateUserPasswordError'
            )
        );
        return rejectWithValue(error);
    }
});

export const authSlice = createSlice({
    name: 'auth',
    initialState,
    reducers: {},
    extraReducers: (builder) => {
        builder.addCase(initialize.fulfilled, (state, action) => {
            state.isInitialized = true;
            state.user = action.payload;
        });
        builder.addCase(initialize.rejected, (state) => {
            state.isInitialized = true;
            state.sessionToken = null;
            state.user = null;
        });

        builder.addCase(login.pending, (state) => {
            state.loginLoading = true;
        });
        builder.addCase(login.fulfilled, (state, action) => {
            state.loginLoading = false;
            state.sessionToken = action.payload.sessionToken;
            state.user = action.payload.user;
        });
        builder.addCase(login.rejected, (state, action) => {
            state.loginLoading = false;
            state.loginError = action.payload;
            state.sessionToken = null;
            state.user = null;
        });

        builder.addCase(logout.pending, (state) => {
            state.loginLoading = false;
            state.loginError = null;
            state.user = null;
        });
        builder.addCase(logout.fulfilled, (state) => {
            state.loginLoading = false;
            state.loginError = null;
            state.sessionToken = null;
            state.user = null;
        });
        builder.addCase(logout.rejected, (state) => {
            state.loginLoading = false;
            state.loginError = null;
            state.sessionToken = null;
            state.user = null;
        });

        builder.addCase(updateExpiredPassword.pending, (state) => {
            state.updateExpiredPasswordLoading = true;
        });
        builder.addCase(updateExpiredPassword.fulfilled, (state, action) => {
            state.updateExpiredPasswordLoading = false;
            state.updateExpiredPasswordError = null;
            state.user = action.payload;
        });
        builder.addCase(updateExpiredPassword.rejected, (state, action) => {
            state.updateExpiredPasswordLoading = false;
            state.updateExpiredPasswordError = action.payload;
        });
    },
});

/**
 * Returns null if the user is not logged in.
 * Otherwise, returns a boolean indicating whether the user's password is expired.
 */
export const authExpiredSelector = createSelector(
    (state: AppState) => state.auth.user,
    (user) => {
        if (user === null) {
            return null;
        }

        return user.AuthSecret !== null && DateTime.fromISO(user.AuthSecret.expires_at) < DateTime.local();
    }
);

/**
 * Returns a boolean indicating whether the user is logged in and does not have an expired password.
 */
export const fullyAuthenticatedSelector = createSelector(
    (state: AppState) => state.auth,
    (authState) => {
        if (authState.user === null || authState.sessionToken === null || authState.isInitialized === false) {
            return false;
        }

        const authExpired =
            authState.user.AuthSecret !== null &&
            DateTime.fromISO(authState.user.AuthSecret.expires_at) < DateTime.local();

        return !authExpired;
    }
);

// Action creators are generated for each case reducer function
// export const {  } = authSlice.actions;

export default authSlice.reducer;
