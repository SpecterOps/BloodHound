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
export class Permission {
    private constructor(private readonly authority: string, private readonly name: string) {
        this.name = name;
        this.authority = authority;
    }

    get() {
        return {
            name: this.name,
            authority: this.authority,
        };
    }

    static readonly APP_READ_APPLICATION_CONFIGURATION = new Permission('app', 'ReadAppConfig');
    static readonly APP_WRITE_APPLICATION_CONFIGURATION = new Permission('app', 'WriteAppConfig');

    static readonly APS_GENERATE_REPORT = new Permission('risks', 'GenerateReport');
    static readonly APS_MANAGE_APS = new Permission('risks', 'ManageRisks');

    static readonly AUTH_ACCEPT_EULA = new Permission('auth', 'AcceptEULA');
    static readonly AUTH_CREATE_TOKEN = new Permission('auth', 'CreateToken');
    static readonly AUTH_MANAGE_APPLICATION_CONFIGURATIONS = new Permission('auth', 'ManageAppConfig');
    static readonly AUTH_MANAGE_PROVIDERS = new Permission('auth', 'ManageProviders');
    static readonly AUTH_MANAGE_SELF = new Permission('auth', 'ManageSelf');
    static readonly AUTH_MANAGE_USERS = new Permission('auth', 'ManageUsers');

    static readonly CLIENTS_MANAGE = new Permission('clients', 'Manage');
    static readonly CLIENTS_READ = new Permission('clients', 'Read');
    static readonly CLIENTS_TASKING = new Permission('clients', 'Tasking');

    static readonly COLLECTION_MANAGE_JOBS = new Permission('collection', 'ManageJobs');

    static readonly GRAPH_DB_READ = new Permission('graphdb', 'Read');
    static readonly GRAPH_DB_WRITE = new Permission('graphdb', 'Write');

    static readonly SAVED_QUERIES_READ = new Permission('saved_queries', 'Read');
    static readonly SAVED_QUERIES_WRITE = new Permission('saved_queries', 'Write');
}
