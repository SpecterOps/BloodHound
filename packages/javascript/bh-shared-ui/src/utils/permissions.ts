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

export enum PermissionsAuthority {
    APP = 'app',
    RISKS = 'risks',
    AUTH = 'auth',
    CLIENTS = 'clients',
    COLLECTION = 'collection',
    GRAPHDB = 'graphdb',
    SAVED_QUERIES = 'saved_queries',
}

export enum PermissionsName {
    READ_APP_CONFIG = 'ReadAppConfig',
    WRITE_APP_CONFIG = 'WriteAppConfig',
    GENERATE_REPORT = 'GenerateReport',
    MANAGE_RISKS = 'ManageRisks',
    CREATE_TOKEN = 'CreateToken',
    MANAGE_APP_CONFIG = 'ManageAppConfig',
    MANAGE_PROVIDERS = 'ManageProviders',
    MANAGE_SELF = 'ManageSelf',
    MANAGE_USERS = 'ManageUsers',
    MANAGE_CLIENTS = 'Manage',
    READ_CLIENTS = 'Read',
    CLIENT_TASKING = 'Tasking',
    MANAGE_COLLECTION_JOBS = 'ManageJobs',
    READ_GRAPHDB = 'Read',
    WRITE_GRAPHDB = 'Write',
    READ_SAVED_QUERIES = 'Read',
    WRITE_SAVED_QUERIES = 'Write',
}

export type PermissionsSpec = {
    authority: PermissionsAuthority;
    name: PermissionsName;
};
