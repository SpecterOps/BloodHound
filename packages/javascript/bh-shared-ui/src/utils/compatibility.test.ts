// Copyright 2024 Specter Ops, Inc.
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

import { isWebGLEnabled } from '.';

const setupMocks = () => {
    const originalRenderingContext = window.WebGLRenderingContext;
    const originalCreateElement = document.createElement;

    const enableWebGLMock = () => {
        window.WebGLRenderingContext = true as any;

        document.createElement = vi.fn((tagName: string) => {
            if (tagName === 'canvas') {
                return {
                    getContext: (contextName: string) => {
                        return !!(contextName === 'webgl');
                    },
                };
            } else {
                return originalCreateElement(tagName);
            }
        }) as any;
    };

    const disableWebGLMock = () => {
        window.WebGLRenderingContext = originalRenderingContext;
        document.createElement = originalCreateElement;
    };

    return { enableWebGLMock, disableWebGLMock };
};

describe('WebGL compatibility check', () => {
    const mocks = setupMocks();
    afterEach(() => mocks.disableWebGLMock());

    it('returns true when WebGL is available in the current browser context', () => {
        mocks.enableWebGLMock();
        expect(isWebGLEnabled()).toBe(true);
    });

    it('returns false when WebGL is not available in the current browser context', () => {
        expect(isWebGLEnabled()).toBe(false);
    });
});
