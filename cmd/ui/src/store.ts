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

import { combineReducers, configureStore, PreloadedState } from '@reduxjs/toolkit';
import { edgeinfo } from 'bh-shared-ui';
import { enableMapSet } from 'immer';
import Cookies from 'js-cookie';
import throttle from 'lodash/throttle';
import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';
import createSagaMiddleware from 'redux-saga';
import * as reducers from 'src/ducks';
import rootSaga from 'src/rootSaga';
import { GlobalViewState } from './ducks/global/types';

enableMapSet();

const sagaMiddleware = createSagaMiddleware();

const appReducer = combineReducers({
    ...reducers,
    edgeinfo,
});

export type RootState = ReturnType<typeof appReducer>;

export const rootReducer = (state: any, action: any): RootState => {
    // If the user logs out, clear the redux store to prevent data leakage
    // Adapted from https://stackoverflow.com/questions/35622588/how-to-reset-the-state-of-a-redux-store
    if (action.type === 'auth/logout/fulfilled' || action.type === 'auth/logout/rejected') {
        const { auth, global } = state;
        state = { auth, global };
        return appReducer(state, action);
    }
    // Otherwise, return the current state
    return appReducer(state, action);
};

const loadState = (): PreloadedState<RootState> => {
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

type PersistedState = {
    auth: { sessionToken: string | null };
    global: {
        view: {
            darkMode: GlobalViewState['darkMode'];
            autoRunQueries: GlobalViewState['autoRunQueries'];
            notifications: GlobalViewState['notifications'];
            exploreLayout: GlobalViewState['exploreLayout'];
        };
    };
};

const saveState = (state: PersistedState) => {
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

const initStore = (preloadedState: PreloadedState<RootState>) => {
    return configureStore({
        reducer: rootReducer,
        preloadedState: preloadedState,
        middleware: (getDefaultMiddleware) => {
            return [...getDefaultMiddleware({ serializableCheck: false }), sagaMiddleware];
        },
    });
};

export const store = initStore(initialState);

// Persist the session token in local storage
store.subscribe(
    throttle(() => {
        const state = store.getState();
        saveState({
            auth: { sessionToken: state.auth.sessionToken },
            global: {
                view: {
                    darkMode: state.global.view.darkMode,
                    autoRunQueries: state.global.view.autoRunQueries,
                    notifications: [],
                    exploreLayout: state.global.view.exploreLayout,
                },
            },
        });
    }, 1000)
);

export type AppState = ReturnType<typeof store.getState>;
export const useAppSelector: TypedUseSelectorHook<AppState> = useSelector;

export type AppDispatch = typeof store.dispatch;
export const useAppDispatch = () => useDispatch<AppDispatch>();

sagaMiddleware.run(rootSaga);
