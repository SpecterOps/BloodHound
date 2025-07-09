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

import { RefObject, useEffect, useState } from 'react';

export function useMeasure(ref: RefObject<HTMLElement>) {
    const [width, setWidth] = useState(0);
    const [height, setHeight] = useState(0);

    useEffect(() => {
        if (!ref.current) return;

        const updateMeasurements = () => {
            if (ref.current) {
                setWidth(ref.current.clientWidth);
                setHeight(ref.current.clientHeight);
            }
        };

        updateMeasurements();
        const observer = new ResizeObserver(updateMeasurements);
        observer.observe(ref.current);

        return () => observer.disconnect();
    }, [ref]);

    return [width, height];
}
