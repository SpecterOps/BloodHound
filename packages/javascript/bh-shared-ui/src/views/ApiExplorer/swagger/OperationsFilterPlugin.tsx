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

import { useEffect, useRef } from 'react';

export const compareTerm = (value: string, term: string): boolean => value.toLowerCase().includes(term.toLowerCase());

/**
 * Filters Swagger tagged operations by a search term.
 *
 * - If a tag name matches the term, all of its operations are included.
 * - If a tag name does not match, only the operations whose path matches the term are included.
 *   The parent tag entry is included with only those matching operations.
 * - Tag entries with no matching tag name or paths are excluded entirely.
 *
 * Matching is case-insensitive and supports partial matches.
 */
export const opsFilter = (taggedOps: any, term: string) => {
    return taggedOps
        .map((tagObj: any, tag: string) => {
            if (compareTerm(tag, term)) {
                return tagObj;
            }
            const matchingOps = tagObj.get('operations').filter((op: any) => compareTerm(op.get('path'), term));
            return tagObj.set('operations', matchingOps);
        })
        .filter((tagObj: any, tag: string) => compareTerm(tag, term) || tagObj.get('operations').size > 0);
};

export const OperationsFilterPlugin = () => {
    return {
        fn: {
            opsFilter,
        },
        wrapComponents: {
            FilterContainer: (Original: any) =>
                function FilterContainer(props: any) {
                    const containerRef = useRef<HTMLDivElement>(null);

                    // Swagger hardcodes the filter input placeholder "Filter by tag"
                    // but our implementation customizes results, included resource path
                    // so this little hack updates the placeholder accordingly
                    useEffect(() => {
                        const input = containerRef.current?.querySelector<HTMLInputElement>('.operation-filter-input');
                        if (input) {
                            input.placeholder = 'Filter by tag or path';
                        }
                    });

                    return (
                        <div ref={containerRef}>
                            <Original {...props} />
                        </div>
                    );
                },
        },
    };
};
