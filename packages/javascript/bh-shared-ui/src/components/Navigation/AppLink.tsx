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

// import { Link, LinkProps, Path } from '@tanstack/react-router';
import { Link, LinkProps, useSearch } from '@tanstack/react-router';

export const AppLink = ({
    children,
    to,
    discardQueryParams,
    ...props
}: LinkProps & { discardQueryParams?: boolean }) => {
    const path = typeof to === 'string' ? to : to.pathname || '';
    const searchParams = useSearch({ strict: false });

    if (discardQueryParams) {
        return (
            <Link to={to} aria-label={`Navigate to ${path}`} {...props}>
                {children}
            </Link>
        );
    }

    return (
        <Link to={to} params={{ search: searchParams }} aria-label={`Navigate to ${path}`} {...props}>
            {children}
        </Link>
    );
};
