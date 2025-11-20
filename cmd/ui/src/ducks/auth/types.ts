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

export interface AuthState {
    isInitialized: boolean;
    loginLoading: boolean;
    loginError: any;
    updateExpiredPasswordLoading: boolean;
    updateExpiredPasswordError: any;
    sessionToken: string | null;
    user: getSelfResponse | null;
}

export interface getSelfResponse {
    id: string;
    principal_name: string;
    email_address: string;
    first_name: string;
    last_name: string;
    saml_provider_id: number | null;
    eula_accepted: boolean;
    last_login: string;
    all_environments: boolean;
    explore_enabled: boolean;
    AuthSecret: {
        id: number;
        digest_method: string;
        expires_at: string;
        totp_activated: boolean;
    };
    roles: {
        id: number;
        name: string;
        description: string;
        permissions: {
            id: number;
            authority: string;
            name: string;
        }[];
    }[];
}
