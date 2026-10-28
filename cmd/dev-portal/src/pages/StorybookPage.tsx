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

import { useEffect, useRef, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { PortalHeader } from '../components/PortalHeader';

const DOODLE_UI_STORYBOOK_URL = 'doodle-ui-storybook/';
const STORY_PATH_PARAM = 'path';
const STORY_PATH_SYNC_INTERVAL_MS = 500;

const buildStorybookUrl = (storyPath: string | null) =>
    storyPath
        ? `${DOODLE_UI_STORYBOOK_URL}?${new URLSearchParams({ [STORY_PATH_PARAM]: storyPath })}`
        : DOODLE_UI_STORYBOOK_URL;

// Storybook changes its ?path= with the History API, which fires no event in the parent page, so the
// same-origin iframe address is read on an interval instead.
const readIframeStoryPath = (iframe: HTMLIFrameElement | null) => {
    try {
        const iframeLocation = iframe?.contentWindow?.location;
        if (!iframeLocation || iframeLocation.href === 'about:blank') return null;
        return new URLSearchParams(iframeLocation.search).get(STORY_PATH_PARAM);
    } catch {
        return null;
    }
};

export function StorybookPage() {
    const iframeRef = useRef<HTMLIFrameElement>(null);
    const [searchParams, setSearchParams] = useSearchParams();
    const storyPath = searchParams.get(STORY_PATH_PARAM);
    const [initialStorybookUrl] = useState(() => buildStorybookUrl(storyPath));

    useEffect(() => {
        const intervalId = window.setInterval(() => {
            const iframeStoryPath = readIframeStoryPath(iframeRef.current);
            if (iframeStoryPath && iframeStoryPath !== storyPath) {
                setSearchParams({ [STORY_PATH_PARAM]: iframeStoryPath }, { replace: true });
            }
        }, STORY_PATH_SYNC_INTERVAL_MS);

        return () => window.clearInterval(intervalId);
    }, [storyPath, setSearchParams]);

    return (
        <div className='portal-frame-shell'>
            <PortalHeader />
            <iframe ref={iframeRef} className='portal-frame' src={initialStorybookUrl} title='Doodle UI Storybook' />
        </div>
    );
}
