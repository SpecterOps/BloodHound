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

export const ThemeEffects = ({ showClippy }: { showClippy: boolean }) => {
    const [dismissed, setDismissed] = useState(false);

    useEffect(() => {
        if (!showClippy) setDismissed(false);
    }, [showClippy]);

    if (!showClippy || dismissed) return null;

    return (
        <aside aria-label='Desktop assistant' className='theme-assistant'>
            <div className='theme-assistant-bubble'>
                <button
                    aria-label='Dismiss desktop assistant'
                    className='theme-assistant-close'
                    onClick={() => setDismissed(true)}
                    type='button'>
                    ×
                </button>
                <strong>Hi! I’m Clippy.</strong>
                <span>It looks like you’re exploring attack paths!</span>
            </div>
            <div aria-hidden='true' className='theme-assistant-character'>
                <span className='theme-assistant-paperclip'>📎</span>
                <span className='theme-assistant-eyes'>● ●</span>
            </div>
        </aside>
    );
};
