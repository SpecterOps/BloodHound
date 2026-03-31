// Copyright 2026 Specter Ops, Inc.
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

import type { ReactNode } from 'react';
import { SubNavSection } from '../../types';
import { XOR } from '../../utils/type';

type MainNavLogoImage = {
    imageUrl: string;
    dimensions: { height: number; width: number };
    classes?: string;
    altText: string;
};

export type MainNavLogoDataObject = {
    project: {
        route: string;
        icon: ReactNode;
    };
    specterOps: {
        image: MainNavLogoImage;
    };
};

type NavItemBase = {
    label: string;
    icon: ReactNode;
    testId: string;
};

type NavActionItem = NavItemBase & {
    control?: ReactNode;
    onClick?: () => void;
};

type NavLinkItem = NavItemBase & {
    route: string;
    target?: React.HTMLAttributeAnchorTarget;
};

type NavSubNavItem = NavItemBase & {
    subNav: SubNavSection[];
};

// XOR is associative so type may be nested for a 3 way exclusive or
export type MainNavDataListItem = XOR<NavActionItem, XOR<NavLinkItem, NavSubNavItem>>;

export type MainNavData = {
    logo: MainNavLogoDataObject;
    primaryList: MainNavDataListItem[];
    secondaryList: MainNavDataListItem[];
};
