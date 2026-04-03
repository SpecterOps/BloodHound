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

import { KeyboardEvent, KeyboardEventHandler } from 'react';

/**
 * Adapts a click handler to a keydown handler by invoking the click handler
 * when the user presses the Enter or Space key.
 *
 * @param handler The click handler to adapt.
 */
export function adaptClickHandlerToKeyDown(handler?: KeyboardEventHandler<HTMLElement>) {
    return (event: KeyboardEvent<HTMLElement>) => {
        if (handler) {
            if (event.key === 'Enter' || event.key === ' ') {
                // Prevent space from scrolling page
                if (event.key === ' ') event.preventDefault();

                handler(event);
            }
        }
    };
}
