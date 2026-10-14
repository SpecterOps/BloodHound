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

import { useCallback, useEffect, useRef, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { PortalHeader } from '../components/PortalHeader';

const DOODLE_UI_ARTIFACT = '/artifacts/doodle-ui/';
const IFRAME_POLL_INTERVAL_MS = 250;

const storybookUrl = (storyPath: string | null) =>
    storyPath ? `${DOODLE_UI_ARTIFACT}?${new URLSearchParams({ path: storyPath })}` : DOODLE_UI_ARTIFACT;

// Returns null while the iframe has no readable Storybook location (no window yet, or cross-origin)
const readStoryPath = (iframe: HTMLIFrameElement | null) => {
    try {
        return new URLSearchParams(iframe?.contentWindow?.location.search ?? '').get('path');
    } catch (error) {
        if (error instanceof DOMException && error.name === 'SecurityError') return null;
        throw error;
    }
};

export const EmbeddedDoodleUIPage = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const portalStoryPath = searchParams.get('path');
    const [initialSrc] = useState(() => storybookUrl(portalStoryPath));
    const iframeRef = useRef<HTMLIFrameElement>(null);
    // The story the portal asked the iframe to show; iframe reads are ignored until it gets there
    const requestedStoryPathRef = useRef<string | null>(null);

    const syncPortalFromIframe = useCallback(() => {
        const iframeStoryPath = readStoryPath(iframeRef.current);
        if (requestedStoryPathRef.current !== null) {
            if (iframeStoryPath === requestedStoryPathRef.current) requestedStoryPathRef.current = null;
            return;
        }
        if (iframeStoryPath && iframeStoryPath !== portalStoryPath) {
            setSearchParams({ path: iframeStoryPath });
        }
    }, [portalStoryPath, setSearchParams]);

    // Storybook changes stories with history.pushState, which fires no load event, so also poll
    useEffect(() => {
        const intervalId = setInterval(syncPortalFromIframe, IFRAME_POLL_INTERVAL_MS);
        return () => clearInterval(intervalId);
    }, [syncPortalFromIframe]);

    // Back/forward change the portal's ?path= first; move the iframe to match
    useEffect(() => {
        const iframeStoryPath = readStoryPath(iframeRef.current);
        const iframeWindow = iframeRef.current?.contentWindow;
        if (iframeWindow && portalStoryPath && iframeStoryPath && iframeStoryPath !== portalStoryPath) {
            requestedStoryPathRef.current = portalStoryPath;
            iframeWindow.location.replace(storybookUrl(portalStoryPath));
        }
    }, [portalStoryPath]);

    return (
        <div className='flex h-screen flex-col'>
            <PortalHeader />
            <iframe
                ref={iframeRef}
                title='DoodleUI Storybook'
                src={initialSrc}
                onLoad={syncPortalFromIframe}
                className='w-full flex-1 border-0'
            />
        </div>
    );
};
