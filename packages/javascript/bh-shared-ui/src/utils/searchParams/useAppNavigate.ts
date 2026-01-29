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

import { NavigateOptions, useNavigate, useSearch } from '@tanstack/react-router';

export const useAppNavigate = () => {
    const navigate = useNavigate();
    const search = useSearch({ strict: false });

    // The navigate() function can optionally take a number as its only argument, which moves up and down the history stack by that amount
    return (to: string, options?: NavigateOptions & { discardQueryParams?: boolean }): void => {
        if (options?.discardQueryParams) {
            navigate({ to, ...options });
        } else {
            navigate({ to, search, ...options });
        }
    };
};
