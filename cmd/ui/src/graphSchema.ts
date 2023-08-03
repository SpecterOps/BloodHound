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

export enum CommonNodeKind {
    MigrationData = 'MigrationData',
}
export function CommonNodeKindToDisplay(value: CommonNodeKind): string | undefined {
    switch (value) {
        case CommonNodeKind.MigrationData:
            return 'MigrationData';
        default:
            return undefined;
    }
}
export enum CommonKind {
    MigrationData = 'MigrationData',
}
export function CommonKindToDisplay(value: CommonKind): string | undefined {
    switch (value) {
        case CommonKind.MigrationData:
            return 'MigrationData';
        default:
            return undefined;
    }
}
export enum CommonKindProperties {
    ObjectID = 'objectid',
    Name = 'name',
    DisplayName = 'displayname',
    Description = 'description',
    OwnerObjectID = 'owner_objectid',
    Collected = 'collected',
    OperatingSystem = 'operatingsystem',
    SystemTags = 'system_tags',
    UserTags = 'user_tags',
    LastSeen = 'lastseen',
    WhenCreated = 'whencreated',
    Enabled = 'enabled',
    PasswordLastSet = 'pwdlastset',
    Title = 'title',
    Email = 'email',
}
export function CommonKindPropertiesToDisplay(value: CommonKindProperties): string | undefined {
    switch (value) {
        case CommonKindProperties.ObjectID:
            return 'Object ID';
        case CommonKindProperties.Name:
            return 'Name';
        case CommonKindProperties.DisplayName:
            return 'Display Name';
        case CommonKindProperties.Description:
            return 'Description';
        case CommonKindProperties.OwnerObjectID:
            return 'Owner Object ID';
        case CommonKindProperties.Collected:
            return 'Collected';
        case CommonKindProperties.OperatingSystem:
            return 'Operating System';
        case CommonKindProperties.SystemTags:
            return 'Node System Tags';
        case CommonKindProperties.UserTags:
            return 'Node User Tags';
        case CommonKindProperties.LastSeen:
            return 'Last Collected by BloodHound';
        case CommonKindProperties.WhenCreated:
            return 'Created';
        case CommonKindProperties.Enabled:
            return 'Enabled';
        case CommonKindProperties.PasswordLastSet:
            return 'Password Last Set';
        case CommonKindProperties.Title:
            return 'Title';
        case CommonKindProperties.Email:
            return 'Email';
        default:
            return undefined;
    }
}
