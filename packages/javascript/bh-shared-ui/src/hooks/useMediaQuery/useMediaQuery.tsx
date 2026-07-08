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
import { useEffect, useState } from 'react';

export const BREAKPOINTS = {
    // tailwind defaults
    sm: '640px',
    md: '768px',
    lg: '1024px',
    xl: '1280px',
    // additions / updates
    '2xl': '1400px',
    '3xl': '1920px',
};

export const useMediaQuery = (query: string): boolean => {
    const [matches, setMatches] = useState(false);

    useEffect(() => {
        const mediaQueryList = window.matchMedia(query);

        // Set initial value
        setMatches(mediaQueryList.matches);

        // Create a listener function
        const documentChangeHandler = (event: MediaQueryListEvent) => {
            setMatches(event.matches);
        };

        // Attach listener
        mediaQueryList.addEventListener('change', documentChangeHandler);

        // Clean up on unmount
        return () => {
            mediaQueryList.removeEventListener('change', documentChangeHandler);
        };
    }, [query]);

    return matches;
};
