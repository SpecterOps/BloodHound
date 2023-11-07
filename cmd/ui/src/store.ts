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
import { useDispatch } from 'react-redux';
import createSagaMiddleware from 'redux-saga';
import * as reducers from 'src/ducks';
import { edgeinfo } from 'bh-shared-ui';
import rootSaga from 'src/rootSaga';

enableMapSet();

const sagaMiddleware = createSagaMiddleware();

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

const reducerManager = createReducerManager({
    ...reducers,
    edgeinfo,
});

const originalStore = configureStore({
    reducer: reducerManager.reduce,
    preloadedState: initialState,
    middleware: (getDefaultMiddleware) => {
        return [...getDefaultMiddleware({ serializableCheck: false }), sagaMiddleware];
    },
});

type BetterStore = typeof originalStore & { reducerManager: ReturnType<typeof createReducerManager> };

export const store: BetterStore = { ...originalStore, reducerManager };

// Persist the session token in local storage
store.subscribe(
    throttle(() => {
        saveState({
            auth: { sessionToken: store.getState().auth.sessionToken },
        });
    }, 1000)
);

export function createReducerManager(initialReducers: any) {
    // Create an object which maps keys to reducers
    const reducers = { ...initialReducers };

    // Create the initial combinedReducer
    let combinedReducer = combineReducers<any, any>(reducers);

    // An array which is used to delete state keys when reducers are removed
    let keysToRemove: any = [];

    return {
        getReducerMap: () => reducers,

        // The root reducer function exposed by this object
        // This will be passed to the store
        reduce: (state: any, action: any) => {
            // If any reducers have been removed, clean up their state first
            if (keysToRemove.length > 0) {
                state = { ...state };
                for (const key of keysToRemove) {
                    delete state[key];
                }
                keysToRemove = [];
            }

            // Delegate to the combined reducer
            return combinedReducer(state, action);
        },

        // Adds a new reducer with the specified key
        add: (key: any, reducer: any) => {
            if (!key || reducers[key]) {
                return;
            }

            // Add the reducer to the reducer mapping
            reducers[key] = reducer;

            // Generate a new combined reducer
            combinedReducer = combineReducers<any, any>(reducers);
        },

        // Removes a reducer with the specified key
        remove: (key: any) => {
            if (!key || !reducers[key]) {
                return;
            }

            // Remove it from the reducer mapping
            delete reducers[key];

            // Add the key to the list of keys to clean up
            keysToRemove.push(key);

            // Generate a new combined reducer
            combinedReducer = combineReducers<any, any>(reducers);
        },

        removeAll: () => {
            combinedReducer = combineReducers<any, any>({ ...initialReducers });
        },
    };
}

export type AppState = ReturnType<typeof store.getState>;

export type AppDispatch = typeof store.dispatch;
export const useAppDispatch = () => useDispatch<AppDispatch>();

sagaMiddleware.run(rootSaga);
