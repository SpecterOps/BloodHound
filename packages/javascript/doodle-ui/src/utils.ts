import { act } from '@testing-library/react';
import { createRoot, Root } from 'react-dom/client';

export const render = (ui: React.ReactElement) => {
    const container = document.createElement('div');
    document.body.appendChild(container);
    let root: Root;
    act(() => {
        root = createRoot(container);
        root.render(ui);
    });
    return {
        container,
        unmount: () => {
            act(() => {
                root.unmount();
            });
            container.remove();
        },
    };
};
