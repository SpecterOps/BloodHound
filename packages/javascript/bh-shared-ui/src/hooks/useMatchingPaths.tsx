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

import { useMatchRoute } from '@tanstack/react-router';

export const useMatchingPaths = (paths: string | string[]) => {
    const matchPath = useMatchRoute();
    if (typeof paths === 'string') {
        const match = matchPath({ to: paths, fuzzy: true });
        return !!match?.pathname;
    } else {
        return paths.reduce(
            (match: boolean, path) => (match ? match : !!matchPath({ to: path, fuzzy: true })?.pathname),
            false
        );
    }
};
