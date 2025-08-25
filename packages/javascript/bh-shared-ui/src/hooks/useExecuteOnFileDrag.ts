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
import { useEffect } from 'react';
import type { AcceptedIngestType } from './useFileIngest';

type ExecuteOnFileDragOptions = {
    acceptedTypes?: AcceptedIngestType[];
    condition?: () => boolean;
};

const alwaysTrue = () => true;

/** Execute the provided function when a file is dragged into the client area */
export const useExecuteOnFileDrag = (
    fn: () => void,
    { acceptedTypes, condition = alwaysTrue }: ExecuteOnFileDragOptions
) => {
    useEffect(() => {
        const onDragEnter = (e: DragEvent) => {
            // Hook only executes when provided condition returns true
            if (!condition()) {
                return;
            }

            if (e.dataTransfer?.types?.includes('Files')) {
                let draggedFiles = Array.from(e.dataTransfer.items);

                // Filter out non-accepted if provided a list
                if (acceptedTypes) {
                    draggedFiles = draggedFiles.filter((item) =>
                        acceptedTypes.includes(item.type as AcceptedIngestType)
                    );
                }

                // Only execute if at least one of the dragged files is accepted
                if (draggedFiles.length) {
                    fn();
                    document.removeEventListener('dragenter', onDragEnter);
                }
            }
        };

        document.addEventListener('dragenter', onDragEnter);

        return () => {
            document.removeEventListener('dragenter', onDragEnter);
        };
    }, [acceptedTypes, condition, fn]);
};
