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

export type BasicResponse<T> = {
    data: T;
};

export type TimeWindowedResponse<T> = BasicResponse<T> & {
    start: string;
    end: string;
};

export type PaginatedResponse<T> = Partial<TimeWindowedResponse<T>> &
    Required<BasicResponse<T>> & {
        count: number;
        limit: number;
        skip: number;
    };

type TimestampFields = {
    created_at: string;
    updated_at: string;
    deleted_at: {
        Time: string;
        Valid: boolean;
    };
};

type PostureStat = TimestampFields & {
    domain_sid: string;
    exposure_index: number;
    tier_zero_count: number;
    critical_risk_count: number;
    id: number;
};

export type PostureResponse = PaginatedResponse<PostureStat[]>;

export type AuthToken = TimestampFields & {
    hmac_method: string;
    id: string;
    last_access: string;
    name: string;
    user_id: string;
};

export type ListAuthTokensResponse = BasicResponse<{ tokens: AuthToken[] }>;

export type NewAuthToken = AuthToken & {
    key: string;
};

export type CreateAuthTokenResponse = BasicResponse<NewAuthToken>;

export type AssetGroupSelector = TimestampFields & {
    id: number;
    asset_group_id: number;
    name: string;
    selector: string;
    system_selector: boolean;
};

export type AssetGroup = TimestampFields & {
    id: number;
    name: string;
    tag: string;
    member_count: number;
    system_group: boolean;
    Selectors: AssetGroupSelector[];
};

export type AssetGroupMember = {
    asset_group_id: number;
    custom_member: boolean;
    environment_id: string;
    environment_kind: string;
    kinds: string[];
    name: string;
    object_id: string;
    primary_kind: string;
};

export type AssetGroupResponse = BasicResponse<{ asset_groups: AssetGroup[] }>;

export type AssetGroupMembersResponse = PaginatedResponse<{ members: AssetGroupMember[] }>;

export type SavedQuery = {
    id: number;
    name: string;
    query: string;
    user_id: string;
};
