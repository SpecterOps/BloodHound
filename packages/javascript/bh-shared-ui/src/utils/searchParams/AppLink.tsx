// Copyright 2025 Specter Ops, Inc.
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

import { Link, LinkProps } from 'react-router-dom';
import {
    AppNavigateProps,
    GloballySupportedSearchParams,
    applyPreservedParams,
    persistSearchParams,
} from './searchParams';

export const AppLink = ({ children, to, discardQueryParams, ...props }: LinkProps & AppNavigateProps) => {
    if (discardQueryParams) {
        return (
            <Link to={to} {...props}>
                {children}
            </Link>
        );
    }

    const search = persistSearchParams(GloballySupportedSearchParams);
    const toWithParams = applyPreservedParams(to, search);

    return (
        <Link to={toWithParams} {...props}>
            {children}
        </Link>
    );
};
