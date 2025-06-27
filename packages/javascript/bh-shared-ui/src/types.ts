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

import { EntityKinds } from './utils/content';

// recursively applies Partial<T> to nested object types

export type DeepPartial<T> = T extends object
    ? {
          [P in keyof T]?: DeepPartial<T[P]>;
      }
    : T;

export type SortOrder = 'asc' | 'desc' | undefined;

export type ValueOf<T> = T[keyof T];

// [key in <string literal>] forces all options in string literal type to be in this map and nothing else
export type MappedStringLiteral<T extends string | number, V = ''> = {
    [key in T]: V;
};

type AdministrationItem = {
    label: string;
    path: string;
    component: React.LazyExoticComponent<React.FC>;
    adminOnly: boolean;
};

export type AdministrationSection = {
    title: string;
    items: AdministrationItem[];
    order: number;
};

export type PrimaryNavItem = {
    label: string;
    icon: JSX.Element;
    route: string;
    testId: string;
};

export type CommonSearchType = {
    subheader: string;
    category: string;
    queries: {
        description: string;
        cypher: string;
    }[];
};

export type SelectedNode = {
    id: string;
    type: EntityKinds;
    name: string;
    graphId?: string;
};

export type BaseGraphLayoutOptions = 'standard' | 'sequential';

export type BaseExploreLayoutOptions = BaseGraphLayoutOptions | 'table';
