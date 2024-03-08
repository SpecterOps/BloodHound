// Copyright 2024 Specter Ops, Inc.
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
export enum Permission {
    APP_READ_APPLICATION_CONFIGURATION,
    APP_WRITE_APPLICATION_CONFIGURATION,
    APS_GENERATE_REPORT,
    APS_MANAGE_APS,
    AUTH_ACCEPT_EULA,
    AUTH_CREATE_TOKEN,
    AUTH_MANAGE_APPLICATION_CONFIGURATIONS,
    AUTH_MANAGE_PROVIDERS,
    AUTH_MANAGE_SELF,
    AUTH_MANAGE_USERS,
    CLIENTS_MANAGE,
    CLIENTS_READ,
    CLIENTS_TASKING,
    COLLECTION_MANAGE_JOBS,
    GRAPH_DB_READ,
    GRAPH_DB_WRITE,
    SAVED_QUERIES_READ,
    SAVED_QUERIES_WRITE,
}

export type PermissionDefinition = {
    authority: string;
    name: string;
};

export type PermissionDefinitions = {
    [index: number]: PermissionDefinition;
};

export const PERMISSIONS: PermissionDefinitions = {
    [Permission.APP_READ_APPLICATION_CONFIGURATION]: {
        authority: 'app',
        name: 'ReadAppConfig',
    },
    [Permission.APP_WRITE_APPLICATION_CONFIGURATION]: {
        authority: 'app',
        name: 'WriteAppConfig',
    },
    [Permission.APS_GENERATE_REPORT]: {
        authority: 'risks',
        name: 'GenerateReport',
    },
    [Permission.APS_MANAGE_APS]: {
        authority: 'risks',
        name: 'ManageRisks',
    },
    [Permission.AUTH_ACCEPT_EULA]: {
        authority: 'auth',
        name: 'AcceptEULA',
    },
    [Permission.AUTH_CREATE_TOKEN]: {
        authority: 'auth',
        name: 'CreateToken',
    },
    [Permission.AUTH_MANAGE_APPLICATION_CONFIGURATIONS]: {
        authority: 'auth',
        name: 'ManageAppConfig',
    },
    [Permission.AUTH_MANAGE_PROVIDERS]: {
        authority: 'auth',
        name: 'ManageProviders',
    },
    [Permission.AUTH_MANAGE_SELF]: {
        authority: 'auth',
        name: 'ManageSelf',
    },
    [Permission.AUTH_MANAGE_USERS]: {
        authority: 'auth',
        name: 'ManageUsers',
    },
    [Permission.CLIENTS_MANAGE]: {
        authority: 'clients',
        name: 'Manage',
    },
    [Permission.CLIENTS_READ]: {
        authority: 'clients',
        name: 'Read',
    },
    [Permission.CLIENTS_TASKING]: {
        authority: 'clients',
        name: 'Tasking',
    },
    [Permission.COLLECTION_MANAGE_JOBS]: {
        authority: 'collection',
        name: 'ManageJobs',
    },
    [Permission.GRAPH_DB_READ]: {
        authority: 'graphdb',
        name: 'Read',
    },
    [Permission.GRAPH_DB_WRITE]: {
        authority: 'graphdb',
        name: 'Write',
    },
    [Permission.SAVED_QUERIES_READ]: {
        authority: 'saved_queries',
        name: 'Read',
    },
    [Permission.SAVED_QUERIES_WRITE]: {
        authority: 'saved_queries',
        name: 'Write',
    },
};
