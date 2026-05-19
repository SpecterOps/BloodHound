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
type AnnouncementSlot = {
    message: string;
    key: number;
};

type AnnouncementState = Record<AnnouncementPriority, AnnouncementSlot>;

const initialState: AnnouncementState = {
    polite: { message: '', key: 0 },
    assertive: { message: '', key: 0 },
};

interface AnnouncementProviderProps {
    children?: ReactNode;
}

const AnnouncementProvider = ({ children }: AnnouncementProviderProps) => {
    const [state, setState] = useState<AnnouncementState>(initialState);

    const announce = useCallback((message: string, priority: AnnouncementPriority = 'polite') => {
        setState((prev) => ({
            ...prev,
            [priority]: {
                message,
                key: prev[priority].key + 1,
            },
        }));
    }, []);

    return (
        <AnnouncementContext.Provider value={announce}>
            {children}
            {(['polite', 'assertive'] as const).map((priority) => (
                <div key={priority} aria-live={priority} aria-atomic='true' className='sr-only'>
                    <span key={state[priority].key}>{state[priority].message}</span>
                </div>
            ))}
        </AnnouncementContext.Provider>
    );
};

export default AnnouncementProvider;
