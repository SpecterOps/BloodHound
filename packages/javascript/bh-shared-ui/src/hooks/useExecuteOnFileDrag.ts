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
