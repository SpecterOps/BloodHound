// Copyright 2023 Specter Ops, Inc.
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

import type { SigmaEdgeEventPayload, SigmaNodeEventPayload } from 'sigma/sigma';
import type { MouseCoords } from 'sigma/types';
import { isNodeEvent, preventAllDefaults } from './utils';

describe('isNodeEvent', () => {
    it('type-gates Sigma node events', () => {
        expect(isNodeEvent({ node: 'node-id' } as SigmaNodeEventPayload)).toBeTruthy();
    });

    it('type-gates Sigma edge events', () => {
        expect(isNodeEvent({ edge: 'edge-id' } as SigmaEdgeEventPayload)).toBeFalsy();
    });
});

describe('preventAllDefaults', () => {
    it('prevents Sigma defaults events', () => {
        const mockEvent: SigmaNodeEventPayload = {
            event: { x: 10, y: 20 } as MouseCoords,
            preventSigmaDefault: vi.fn(),
            node: 'node-id',
        };

        preventAllDefaults(mockEvent);
        expect(mockEvent.preventSigmaDefault).toHaveBeenCalled();
    });

    it('prevents MouseCoords events', () => {
        const mockEvent = new MouseEvent('click');
        mockEvent.preventDefault = vi.fn();
        mockEvent.stopPropagation = vi.fn();
        const mockMouseCoords: MouseCoords = {
            sigmaDefaultPrevented: false,
            preventSigmaDefault: vi.fn(),
            original: mockEvent,
            x: 10,
            y: 20,
        };

        preventAllDefaults(mockMouseCoords);
        expect(mockMouseCoords.preventSigmaDefault).toHaveBeenCalled();
        expect(mockMouseCoords.original.preventDefault).toHaveBeenCalled();
        expect(mockMouseCoords.original.stopPropagation).toHaveBeenCalled();
    });
});
