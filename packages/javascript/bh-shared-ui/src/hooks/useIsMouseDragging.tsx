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

import { useEffect, useRef, useState } from 'react';

export const useIsMouseDragging = () => {
    const [isMouseDragging, setIsMouseDragging] = useState<boolean>(false);
    const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    const handlePointerDown = () => {
        // We are setting a timeout here so that state is not changed unless the user holds down the mouse button for some period of time
        timeoutRef.current = setTimeout(() => {
            setIsMouseDragging(true);
        }, 200);
    };

    const handlePointerUp = () => {
        setIsMouseDragging(false);

        if (timeoutRef.current) {
            clearTimeout(timeoutRef.current);
        }
    };

    useEffect(() => {
        document.addEventListener('mousedown', handlePointerDown);
        document.addEventListener('mouseup', handlePointerUp);

        return () => {
            document.removeEventListener('mousedown', handlePointerDown);
            document.removeEventListener('mouseup', handlePointerUp);
        };
    }, []);

    return { isMouseDragging };
};
