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

import { ReactNode, createContext, useCallback, useState } from 'react';

export type AnnouncementPriority = 'polite' | 'assertive';

export type AnnounceFunction = (message: string, priority?: AnnouncementPriority) => void;

export const AnnouncementContext = createContext<AnnounceFunction | null>(null);

// Each priority uses a single permanently-mounted aria-live div with a keyed
// child <span> inside it. The outer div stays in the DOM so screen readers keep
// it established in the accessibility tree. On every announce call the key
// counter increments, which causes React to unmount the old <span> and mount a
// new one — a DOM mutation the live region detects and announces. This handles
// re-announcing the same message without needing multiple slots or setTimeout.
interface AnnouncementProviderProps {
    children?: ReactNode;
}

const AnnouncementProvider = ({ children }: AnnouncementProviderProps) => {
    const [polite, setPolite] = useState({ message: '', key: 0 });
    const [assertive, setAssertive] = useState({ message: '', key: 0 });

    const announce = useCallback((message: string, priority: AnnouncementPriority = 'polite') => {
        const setter = priority === 'polite' ? setPolite : setAssertive;
        setter((prev) => ({ message, key: prev.key + 1 }));
    }, []);

    const slots = { polite, assertive };

    return (
        <AnnouncementContext.Provider value={announce}>
            {children}
            {(['polite', 'assertive'] as const).map((priority) => (
                <div key={priority} aria-live={priority} aria-atomic='true' className='sr-only'>
                    <span key={slots[priority].key}>{slots[priority].message}</span>
                </div>
            ))}
        </AnnouncementContext.Provider>
    );
};

export default AnnouncementProvider;
