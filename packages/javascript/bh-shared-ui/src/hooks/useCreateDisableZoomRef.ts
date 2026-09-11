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
import { useCallback } from 'react';

/*
 *  This function returns a callback to be applied to the ref={} prop of an HTML element
 *  which you want to disable browser magnifying zoom (in our case, the explore graph).
 *  This function accepts a callback to be applied to the wheel handler.
 */
const useCreateDisableZoomRef = <T extends HTMLElement>(onWheel?: (e: WheelEvent) => void) => {
    return useCallback(
        (ref: T) => {
            // Clean up existing listener if any
            ref?.removeEventListener('wheel', (ref as any)._wheelHandler as any);

            const wheelHandler = (e: WheelEvent) => {
                e.preventDefault();
                if (typeof onWheel === 'function') {
                    onWheel(e);
                }
            };

            if (ref) {
                // Store reference for cleanup
                (ref as any)._wheelHandler = wheelHandler;

                ref?.addEventListener('wheel', wheelHandler, { passive: false });
            }
        },
        [onWheel]
    );
};

export default useCreateDisableZoomRef;
