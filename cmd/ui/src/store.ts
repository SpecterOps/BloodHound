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

import { combineReducers, configureStore } from '@reduxjs/toolkit';
import { enableMapSet } from 'immer';
import Cookies from 'js-cookie';
import throttle from 'lodash/throttle';
import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';
import createSagaMiddleware from 'redux-saga';
import * as reducers from 'src/ducks';
import { edgeinfo } from 'bh-shared-ui';
import rootSaga from 'src/rootSaga';

enableMapSet();

const sagaMiddleware = createSagaMiddleware();

const appReducer = combineReducers({
    ...reducers,
    edgeinfo,
});

export const rootReducer = (state: any, action: any) => {
    // If the user logs out, clear the redux store to prevent data leakage
    // Adapted from https://stackoverflow.com/questions/35622588/how-to-reset-the-state-of-a-redux-store
    if (action.type === 'auth/logout/fulfilled' || action.type === 'auth/logout/rejected') {
        const { auth } = state;
        state = { auth };
        return appReducer(state, action);
    }

    // Otherwise, return the current state
    return appReducer(state, action);
};

const loadState = () => {
    try {
        const serializedState = localStorage.getItem('persistedState');
        if (serializedState === null) {
            return {};
        }
        return JSON.parse(serializedState);
    } catch (error) {
        return {};
    }
};

const saveState = (state: any) => {
    try {
        const serializedState = JSON.stringify(state);
        localStorage.setItem('persistedState', serializedState);
    } catch (error) {
        // Ignore
    }
};

const initialState = loadState();

// If this is a SAML login, we need to override any persisted session token
const SAMLToken = Cookies.get('token');
if (SAMLToken !== undefined) {
    initialState.auth = Object.assign({}, initialState.auth, { sessionToken: SAMLToken });
    Cookies.remove('token');
}

export const store = configureStore({
    reducer: rootReducer,
    preloadedState: initialState,
    middleware: (getDefaultMiddleware) => {
        return [...getDefaultMiddleware({ serializableCheck: false }), sagaMiddleware];
    },
});

// Persist the session token in local storage
store.subscribe(
    throttle(() => {
        saveState({
            auth: { sessionToken: store.getState().auth.sessionToken },
        });
    }, 1000)
);

export type AppState = ReturnType<typeof store.getState>;
export const useAppSelector: TypedUseSelectorHook<AppState> = useSelector

export type AppDispatch = typeof store.dispatch;
export const useAppDispatch = () => useDispatch<AppDispatch>();

sagaMiddleware.run(rootSaga);
