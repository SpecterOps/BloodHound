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
import { useMemo } from 'react';
import useDebouncedValue from '../hooks/useDebouncedValue';

/* This helper component is used for screen readers to have a debounced
announcement when typing into a filter or text field as the results are announced. 
This helps with the issue of the screen reader cutting itself off
while still announcing other relevant info the user needs to hear. */

export const ANNOUNCEMENT_DEBOUNCE_MS = 1000;

type DebouncedAnnouncementProps = {
    // Change the key when the user's query changes, even if the message is identical.
    announcementKey: string;
    message: string;
    enabled?: boolean;
    delay?: number;
};

export const DebouncedAnnouncement = ({
    announcementKey,
    message,
    enabled = true,
    delay = ANNOUNCEMENT_DEBOUNCE_MS,
}: DebouncedAnnouncementProps) => {
    const announcement = useMemo(
        () => ({ key: announcementKey, message: enabled ? message : '' }),
        [announcementKey, message, enabled]
    );
    // Let keyboard echo finish without delaying the UI that produced the message.
    const debouncedAnnouncement = useDebouncedValue(announcement, delay);

    return (
        <div className='sr-only' aria-live='polite' aria-atomic='true' role='status'>
            {enabled && debouncedAnnouncement.message && (
                // Keep the live region mounted and replace its child for equal-count queries.
                <span key={debouncedAnnouncement.key}>{debouncedAnnouncement.message}</span>
            )}
        </div>
    );
};
