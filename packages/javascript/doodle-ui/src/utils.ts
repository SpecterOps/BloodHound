import { createRoot } from 'react-dom/client';
import { act } from 'react-dom/test-utils';

export const render = (ui: React.ReactElement) => {
    const container = document.createElement('div');
    document.body.appendChild(container);
    act(() => {
        createRoot(container).render(ui);
    });
    return container;
};
